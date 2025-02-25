package internal

import (
	"errors"
	"fmt"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/lib"
	"github.com/busy-cloud/boat/mqtt"
	"sync"
	"time"
)

var gateways lib.Map[ModbusMaster]

type ModbusMaster struct {
	Id         string    `json:"id" xorm:"pk"`
	Name       string    `json:"name,omitempty"`
	LinkerId   string    `json:"linker_id" xorm:"index"`
	IncomingId string    `json:"incoming_id" xorm:"index"`
	Timeout    int64     `json:"timeout"` //超时
	Disabled   bool      `json:"disabled,omitempty"`
	Created    time.Time `json:"created,omitempty" xorm:"created"`

	//packets chan *Packet
	devices map[string]*Device

	opened bool

	wait chan []byte
}

func (g *ModbusMaster) Write(request []byte) error {
	tkn := mqtt.Publish("link/"+g.LinkerId+"/"+g.IncomingId+"/down", request)
	tkn.Wait()
	return tkn.Error()
}

func (g *ModbusMaster) Read() ([]byte, error) {
	select {
	case buf := <-g.wait:
		return buf, nil
	case <-time.After(time.Second * time.Duration(g.Timeout)):
		return nil, errors.New("timeout")
	}
}

func (g *ModbusMaster) ReadAtLeast(n int) ([]byte, error) {
	var ret []byte

	for len(ret) < n {
		buf, err := g.Read()
		if err != nil {
			return nil, err
		}
		ret = append(ret, buf...)
	}

	return ret, nil
}

func (g *ModbusMaster) onData(buf []byte) {
	g.wait <- buf
}

func (g *ModbusMaster) Close() error {
	if !g.opened {
		return fmt.Errorf("gateway already closed")
	}
	g.opened = false
	return nil
}

func (g *ModbusMaster) Open() error {
	if g.opened {
		return fmt.Errorf("gateway is already opened")
	}

	err := g.LoadDevices()
	if err != nil {
		return err
	}

	g.opened = true

	go g.polling()

	return nil
}

func (g *ModbusMaster) polling() {
	for g.opened {
		//TODO 轮询

	}
}

func (g *ModbusMaster) LoadDevice(id string) error {
	var device Device
	has, err := db.Engine.ID(id).Get(&device)
	if err != nil {
		return err
	}
	if !has {
		return fmt.Errorf("device %s not found", id)
	}
	g.devices[id] = &device
	return nil
}

func (g *ModbusMaster) UnLoadDevice(id string) {
	delete(g.devices, id)
}

func (g *ModbusMaster) LoadDevices() error {
	//清空
	g.devices = make(map[string]*Device)

	var devices []*Device
	err := db.Engine.Where("gateway_id=?", g.Id).Find(&devices)
	if err != nil {
		return err
	}
	for _, device := range devices {
		g.devices[device.Id] = device
		device.gateway = g
	}
	return nil
}

func (g *ModbusMaster) GetDevice(id string) *Device {
	return g.devices[id]
}

func LoadGateway(id string) (*ModbusMaster, error) {
	var gateway ModbusMaster
	has, err := db.Engine.ID(id).Get(&gateway)
	if err != nil {
		return nil, err
	}
	//数据库里查不到，则创建
	if !has {
		gateway.Id = id
		_, err := db.Engine.Insert(&gateway)
		if err != nil {
			return nil, err
		}
	}

	//加载，即打开
	//err = gateway.Open()
	//if err != nil {
	//	return nil, err
	//}

	return &gateway, nil
}

var ensureLock sync.Mutex

// 自动加载网关
func EnsureGateway(id string) (gateway *ModbusMaster, err error) {
	//此处应该加锁，避免重复创建
	ensureLock.Lock()
	defer ensureLock.Unlock()

	gateway = gateways.Load(id)
	if gateway == nil {
		gateway, err = LoadGateway(id)
		if err != nil {
			return nil, err
		}

		gateways.Store(id, gateway)
	}
	return
}

func GetGateway(id string) *ModbusMaster {
	return gateways.Load(id)
}
