package internal

import (
	"errors"
	"fmt"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/lib"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"time"
)

var masters lib.Map[Master]

type Master struct {
	LinkerId   string `json:"linker_id" xorm:"index"`
	IncomingId string `json:"incoming_id" xorm:"index"`

	//packets chan *Packet
	devices map[string]*Device

	opened bool

	wait chan []byte
}

func (g *Master) Write(request []byte) error {
	tkn := mqtt.Publish("link/"+g.LinkerId+"/"+g.IncomingId+"/down", request)
	tkn.Wait()
	return tkn.Error()
}

func (g *Master) Read() ([]byte, error) {
	select {
	case buf := <-g.wait:
		return buf, nil
	case <-time.After(time.Second * 5):
		return nil, errors.New("timeout")
	}
}

func (g *Master) ReadAtLeast(n int) ([]byte, error) {
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

func (g *Master) onData(buf []byte) {
	g.wait <- buf
}

func (g *Master) Close() error {
	if !g.opened {
		return fmt.Errorf("master already closed")
	}
	g.opened = false
	return nil
}

func (g *Master) Open() error {
	if g.opened {
		return fmt.Errorf("master is already opened")
	}

	err := g.LoadDevices()
	if err != nil {
		return err
	}

	g.opened = true

	go g.polling()

	return nil
}

func (g *Master) polling() {
	for g.opened {
		//TODO 轮询

	}
}

func (g *Master) LoadDevice(id string) error {
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

func (g *Master) UnLoadDevice(id string) {
	delete(g.devices, id)
}

func (g *Master) LoadDevices() error {
	//清空
	g.devices = make(map[string]*Device)

	var devices []*Device
	err := db.Engine.Where("linker_id=?", g.LinkerId).And("incoming_id=?", g.IncomingId).Find(&devices)
	if err != nil {
		return err
	}
	for _, device := range devices {
		g.devices[device.Id] = device
		device.master = g
		device.product, err = EnsureProduct(device.ProductId)
		if err != nil {
			log.Printf("failed to ensure product: %v", err)
		}
	}
	return nil
}

func (g *Master) GetDevice(id string) *Device {
	return g.devices[id]
}

// 自动加载网关
func EnsureMaster(linker, incoming string) (master *Master, err error) {
	//此处应该加锁，避免重复创建

	id := linker + "/" + incoming
	master = masters.Load(id)
	if master == nil {
		master = &Master{
			LinkerId:   linker,
			IncomingId: incoming,
			wait:       make(chan []byte),
		}

		masters.Store(id, master)
	}
	return
}

func GetMaster(linker, incoming string) *Master {
	id := linker + "/" + incoming
	return masters.Load(id)
}
