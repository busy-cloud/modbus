package internal

import (
	"errors"
	"fmt"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/busy-cloud/boat/pool"
	"sync"
	"time"
)

// ModbusMaster modbus主站
type ModbusMaster struct {
	Id              string    `json:"id,omitempty" xorm:"pk"`
	Name            string    `json:"name,omitempty"`
	Description     string    `json:"description,omitempty"`
	LinkerId        string    `json:"linker_id" xorm:"index"`
	IncomingId      string    `json:"incoming_id" xorm:"index"`
	Polling         bool      `json:"polling,omitempty"`          //开启轮询
	PollingInterval uint      `json:"polling_interval,omitempty"` //轮询间隔(s)
	Disabled        bool      `json:"disabled,omitempty"`         //禁用
	Created         time.Time `json:"created,omitempty" xorm:"created"`

	//packets chan *Packet
	devices map[string]*Device

	opened bool

	wait chan []byte
	lock sync.Mutex
}

func (m *ModbusMaster) LinkerAndIncomingID() string {
	return m.LinkerId + "_" + m.IncomingId
}

func (m *ModbusMaster) Write(request []byte) error {
	return WriteTo(m.LinkerId, m.IncomingId, request)
}

func (m *ModbusMaster) Read() ([]byte, error) {
	select {
	case buf := <-m.wait:
		return buf, nil
	case <-time.After(time.Second * 5):
		return nil, errors.New("timeout")
	}
}

func (m *ModbusMaster) ReadAtLeast(n int) ([]byte, error) {
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

func (m *ModbusMaster) Ask(request []byte, n int) ([]byte, error) {
	//加锁，避免重入（同一连接下，线程均等待，回头可以改成队列）
	m.lock.Lock()
	defer m.lock.Unlock()

	//发送请求
	if len(request) > 0 {
		err := m.Write(request)
		if err != nil {
			return nil, err
		}
	}

	var ret []byte

	for len(ret) < n {
		buf, err := m.Read()
		if err != nil {
			return nil, err
		}
		ret = append(ret, buf...)
		if len(ret) > 2 {
			if ret[1]&0x80 > 0 {
				return nil, fmt.Errorf("modbus error %d", ret[1])
			}
		}
	}

	return ret, nil
}

func (m *ModbusMaster) onData(buf []byte) {
	m.wait <- buf
}

func (m *ModbusMaster) Close() error {
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

func (m *ModbusMaster) Open() error {
	if m.opened {
		return fmt.Errorf("master is already opened")
	}

	m.wait = make(chan []byte)

	err := m.LoadDevices()
	if err != nil {
		return err
	}

	m.opened = true

	if m.Polling {
		go m.polling()
	}

	return nil
}

func (m *ModbusMaster) LoadDevice(id string) error {
	var device Device
	has, err := db.Engine().ID(id).Get(&device)
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

func (m *ModbusMaster) UnLoadDevice(id string) {
	if d, ok := m.devices[id]; ok {
		_ = d.Close()
		delete(m.devices, id)
	}
}

func (m *ModbusMaster) LoadDevices() error {
	//清空
	m.devices = make(map[string]*Device)

	var devices []*Device
	err := db.Engine().Where("master_id=?", m.Id).Find(&devices)
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

func (m *ModbusMaster) GetDevice(id string) *Device {
	return m.devices[id]
}

func (m *ModbusMaster) polling() {
	for m.opened {
		for _, device := range m.devices {

			//异步读
			_ = pool.Insert(func() {
				values, err := device.Poll()
				if err != nil {
					log.Error(err)
					return
				}
				topic := fmt.Sprintf("device/%s/%s/property", device.ProductId, device.Id)
				mqtt.Publish(topic, values)
			})

			//加上小间隔
			//time.Sleep(1 * time.Second)
		}

		//轮询间隔
		if m.PollingInterval > 0 {
			time.Sleep(time.Duration(m.PollingInterval) * time.Second)
		} else {
			time.Sleep(time.Minute)
		}
	}
}
