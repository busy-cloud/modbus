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

func (m *Master) Write(request []byte) error {
	tkn := mqtt.Publish("link/"+m.LinkerId+"/"+m.IncomingId+"/down", request)
	tkn.Wait()
	return tkn.Error()
}

func (m *Master) Read() ([]byte, error) {
	select {
	case buf := <-m.wait:
		return buf, nil
	case <-time.After(time.Second * 5):
		return nil, errors.New("timeout")
	}
}

func (m *Master) ReadAtLeast(n int) ([]byte, error) {
	var ret []byte

	for len(ret) < n {
		buf, err := m.Read()
		if err != nil {
			return nil, err
		}
		ret = append(ret, buf...)
	}

	return ret, nil
}

func (m *Master) onData(buf []byte) {
	m.wait <- buf
}

func (m *Master) Close() error {
	if !m.opened {
		return fmt.Errorf("master already closed")
	}
	m.opened = false

	for _, device := range m.devices {
		_ = device.Close()
	}
	m.devices = nil
	close(m.wait)

	return nil
}

func (m *Master) Open() error {
	if m.opened {
		return fmt.Errorf("master is already opened")
	}

	m.wait = make(chan []byte)

	err := m.LoadDevices()
	if err != nil {
		return err
	}

	m.opened = true

	return nil
}

func (m *Master) LoadDevice(id string) error {
	var device Device
	has, err := db.Engine.ID(id).Get(&device)
	if err != nil {
		return err
	}
	if !has {
		return fmt.Errorf("device %s not found", id)
	}
	m.devices[id] = &device
	device.master = m
	device.product, err = EnsureProduct(device.ProductId)
	if err != nil {
		log.Printf("failed to ensure product: %v", err)
	}
	err = device.Open()
	if err != nil {
		log.Printf("failed to open device: %v", err)
	}
	return nil
}

func (m *Master) UnLoadDevice(id string) {
	if d, ok := m.devices[id]; ok {
		_ = d.Close()
		delete(m.devices, id)
	}
}

func (m *Master) LoadDevices() error {
	//清空
	m.devices = make(map[string]*Device)

	var devices []*Device
	err := db.Engine.Where("linker_id=?", m.LinkerId).And("incoming_id=?", m.IncomingId).Find(&devices)
	if err != nil {
		return err
	}
	for _, device := range devices {
		m.devices[device.Id] = device
		device.master = m
		device.product, err = EnsureProduct(device.ProductId)
		if err != nil {
			log.Printf("failed to ensure product: %v", err)
		}
		err = device.Open()
		if err != nil {
			log.Printf("failed to open device: %v", err)
		}
	}
	return nil
}

func (m *Master) GetDevice(id string) *Device {
	return m.devices[id]
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
			//wait:       make(chan []byte),
		}

		masters.Store(id, master)

		err = master.Open()
		//if err != nil {
		//	return nil, err
		//}
	}
	return
}

func GetMaster(linker, incoming string) *Master {
	id := linker + "/" + incoming
	return masters.Load(id)
}
