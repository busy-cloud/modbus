package internal

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/spf13/cast"
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

func (m *ModbusMaster) Write(slave, code uint8, offset uint16, value any) error {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(slave)
	buf.WriteByte(code)
	_ = binary.Write(buf, binary.BigEndian, offset)
	switch code {
	case 5: //单个线圈
		if cast.ToBool(value) {
			buf.WriteByte(0xff)
			buf.WriteByte(0xff)
		} else {
			buf.WriteByte(0x00)
			buf.WriteByte(0xff)
		}
	//case 15: //多个线圈
	case 6: //单个寄存器
		_ = binary.Write(buf, binary.BigEndian, cast.ToUint16(value))
	//case 16: //多个寄存器
	default:
		return fmt.Errorf("invalid code: %d", code)
	}

	_ = binary.Write(buf, binary.LittleEndian, CRC16(buf.Bytes()))

	//发送
	_, err := m.ask(buf.Bytes(), buf.Len()) //写数据时，返回数据一样，长度也一样

	return err
}

func (m *ModbusMaster) Read(slave, code uint8, offset uint16, length uint16) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(slave)
	buf.WriteByte(code)
	_ = binary.Write(buf, binary.BigEndian, offset)
	_ = binary.Write(buf, binary.BigEndian, length)
	_ = binary.Write(buf, binary.LittleEndian, CRC16(buf.Bytes()))

	//发送
	res, err := m.ask(buf.Bytes(), 7)
	if err != nil {
		return nil, err
	}

	cnt := int(res[2]) //字节数
	ln := 5 + cnt
	//长度不够，继续读
	if len(res) < ln {
		b, e := m.ask(nil, ln-len(res))
		if e != nil {
			return nil, e
		}
		res = append(res, b...)
	}

	return res[3 : len(res)-2], nil //除去包头和crc校验码
}

func (m *ModbusMaster) write(request []byte) error {
	return WriteTo(m.LinkerId, m.IncomingId, request)
}

func (m *ModbusMaster) read() ([]byte, error) {
	select {
	case buf := <-m.wait:
		return buf, nil
	case <-time.After(time.Second * 5):
		return nil, errors.New("timeout")
	}
}

func (m *ModbusMaster) readAtLeast(n int) ([]byte, error) {
	var ret []byte

	for len(ret) < n {
		buf, err := m.read()
		if err != nil {
			return nil, err
		}
		ret = append(ret, buf...)
	}

	return ret, nil
}

func (m *ModbusMaster) ask(request []byte, n int) ([]byte, error) {
	//加锁，避免重入（同一连接下，线程均等待，回头可以改成队列）
	m.lock.Lock()
	defer m.lock.Unlock()

	//发送请求
	if len(request) > 0 {
		err := m.write(request)
		if err != nil {
			return nil, err
		}
	}

	var ret []byte

	for len(ret) < n {
		buf, err := m.read()
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
			//_ = pool.Insert(func() {
			values, err := device.Poll()
			if err != nil {
				log.Error(err)
				return
			}
			topic := fmt.Sprintf("device/%s/%s/property", device.ProductId, device.Id)
			mqtt.Publish(topic, values)
			//})

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
