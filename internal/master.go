package internal

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/god-jason/iot-master/bin"
	"github.com/god-jason/iot-master/protocol"
	"go.uber.org/multierr"
	"sync"
	"sync/atomic"
	"time"
)

// ModbusMaster modbus主站
type ModbusMaster struct {
	*Options

	//Id         string
	Linker string
	LinkId string
	writer protocol.WriteLinkFunc

	//packets chan *Packet
	devices map[string]*Device

	opened bool

	wait    chan []byte
	waiting atomic.Bool

	//读写事务锁，避免重入
	lock sync.Mutex

	//tcp自增ID
	increment uint16
}

func (m *ModbusMaster) Write(slave, code uint8, offset uint16, value []byte) error {
	buf := bytes.NewBuffer(nil)
	if m.Tcp {
		_ = binary.Write(buf, binary.BigEndian, m.increment)
		m.increment++
		_ = binary.Write(buf, binary.BigEndian, 0)
		_ = binary.Write(buf, binary.BigEndian, 0)
	}
	buf.WriteByte(slave)
	buf.WriteByte(code)
	_ = binary.Write(buf, binary.BigEndian, offset)
	buf.Write(value)
	if !m.Tcp {
		_ = binary.Write(buf, binary.LittleEndian, CRC16(buf.Bytes()))
	}

	//发送
	_, err := m.ask(buf.Bytes(), buf.Len()) //写数据时，返回数据一样，长度也一样

	//TODO 判断错误码

	return err
}

func (m *ModbusMaster) Read(slave, code uint8, offset uint16, length uint16) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if m.Tcp {
		_ = binary.Write(buf, binary.BigEndian, m.increment)
		m.increment++
		_ = binary.Write(buf, binary.BigEndian, 0)
		_ = binary.Write(buf, binary.BigEndian, 0)
	}
	buf.WriteByte(slave)
	buf.WriteByte(code)
	_ = binary.Write(buf, binary.BigEndian, offset)
	_ = binary.Write(buf, binary.BigEndian, length)
	if !m.Tcp {
		_ = binary.Write(buf, binary.LittleEndian, CRC16(buf.Bytes()))
	}

	want := 7
	if m.Tcp {
		want = 8
	}

	//发送
	res, err := m.ask(buf.Bytes(), want)
	if err != nil {
		return nil, err
	}

	ln := 0
	if m.Tcp {
		remain := bin.ParseUint16(res[4:])
		ln = int(remain) + 4

		//判断错误码
		if res[7] > 0x80 {
			return nil, fmt.Errorf("invalid code: %d", res[7])
		}
	} else {
		//计算字节数
		cnt := int(res[2])
		ln = 5 + cnt

		//判断错误码
		if res[1] > 0x80 {
			return nil, fmt.Errorf("invalid code: %d", res[1])
		}
	}

	//长度不够，继续读
	if len(res) < ln {
		b, e := m.ask(nil, ln-len(res))
		if e != nil {
			return nil, e
		}
		res = append(res, b...)
	}

	if m.Tcp {
		return res[9:], nil //去掉包头
	} else {
		return res[3 : len(res)-2], nil //除去包头和crc校验码
	}
}

func (m *ModbusMaster) write(request []byte) error {
	return m.writer(m.Linker, m.LinkId, request)
}

func (m *ModbusMaster) read() ([]byte, error) {
	m.waiting.Store(true)
	defer m.waiting.Store(false)

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

func (m *ModbusMaster) OnData(buf []byte) {
	//此处判断是否有等待
	if m.waiting.Load() {
		m.wait <- buf
	}
}

func (m *ModbusMaster) Close() error {
	if !m.opened {
		return fmt.Errorf("master already closed")
	}
	m.opened = false

	//m.devices = nil
	close(m.wait)

	return nil
}

func (m *ModbusMaster) Open() error {
	if m.opened {
		return fmt.Errorf("master is already opened")
	}

	m.devices = make(map[string]*Device)
	m.wait = make(chan []byte)

	m.opened = true

	if m.Polling {
		go m.polling()
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
				continue
			}
			topic := fmt.Sprintf("device/%s/values", device.Id)
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

func (m *ModbusMaster) OnSync(request *protocol.SyncRequest) (*protocol.SyncResponse, error) {
	dev, ok := m.devices[request.DeviceId]
	if !ok {
		return nil, fmt.Errorf("device %s not found", request.DeviceId)
	}
	values, err := dev.Poll()
	if err != nil {
		return nil, err
	}
	return &protocol.SyncResponse{
		MsgId:    request.MsgId,
		DeviceId: request.DeviceId,
		Values:   values,
	}, nil
}

func (m *ModbusMaster) OnRead(request *protocol.ReadRequest) (*protocol.ReadResponse, error) {
	dev, ok := m.devices[request.DeviceId]
	if !ok {
		return nil, fmt.Errorf("device %s not found", request.DeviceId)
	}

	var e error
	resp := &protocol.ReadResponse{
		MsgId:    request.MsgId,
		DeviceId: request.DeviceId,
		Values:   make(map[string]any),
	}

	for _, point := range request.Points {
		val, err := dev.Get(point)
		if err != nil {
			e = multierr.Append(e, err)
			continue
		}
		resp.Values[point] = val
	}

	if e != nil {
		if len(resp.Values) == 0 {
			return nil, e
		}
		//有成功有失败
		resp.Error = e.Error()
	}

	return resp, nil
}

func (m *ModbusMaster) OnWrite(request *protocol.WriteRequest) (*protocol.WriteResponse, error) {
	dev, ok := m.devices[request.DeviceId]
	if !ok {
		return nil, fmt.Errorf("device %s not found", request.DeviceId)
	}

	var e error
	resp := &protocol.WriteResponse{
		MsgId:    request.MsgId,
		DeviceId: request.DeviceId,
		Result:   make(map[string]bool),
	}

	for point, value := range request.Values {
		err := dev.Set(point, value)
		if err != nil {
			e = multierr.Append(e, err)
			continue
		}

		resp.Result[point] = true
	}

	if e != nil {
		if len(resp.Result) == 0 {
			return nil, e
		}
		//有成功有失败
		resp.Error = e.Error()
	}

	return resp, nil
}

func (m *ModbusMaster) OnAction(request *protocol.ActionRequest) (*protocol.ActionResponse, error) {
	dev, ok := m.devices[request.DeviceId]
	if !ok {
		return nil, fmt.Errorf("device %s not found", request.DeviceId)
	}

	resp := &protocol.ActionResponse{
		MsgId:    request.MsgId,
		DeviceId: request.DeviceId,
	}

	err := dev.Action(request.Action, request.Parameters)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (m *ModbusMaster) OnAttach(payload []byte) {
	var devs []*Device
	err := json.Unmarshal(payload, &devs)
	if err != nil {
		log.Error(err)
		return
	}

	for _, dev := range devs {
		dev.master = m
		m.devices[dev.Id] = dev
	}
}

func (m *ModbusMaster) OnDetach(payload []byte) {
	var devs []string
	err := json.Unmarshal(payload, &devs)
	if err != nil {
		log.Error(err)
		return
	}
	for _, dev := range devs {
		delete(m.devices, dev)
	}
}
