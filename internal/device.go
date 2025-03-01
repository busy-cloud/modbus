package internal

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/busy-cloud/boat/cron"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/busy-cloud/iot/types"
	"github.com/spf13/cast"
	"go.uber.org/multierr"
)

type Device struct {
	types.Device `xorm:"extends"`

	MasterId string `json:"master_id,omitempty" xorm:"index"` //主站ID
	Slave    uint8  `json:"slave,omitempty"`                  //从站号

	master  *ModbusMaster
	product *Product
	jobs    []*cron.Job
}

func (d *Device) Open() error {
	if d.product == nil {
		return errors.New("product not exist")
	}
	if d.product.pollers == nil {
		return errors.New("product.pollers not exist")
	}

	p := d.product.pollers

	fn := func() {
		values, err := d.Poll()
		if err != nil {
			log.Error(err)
			return
		}

		if len(values) > 0 {
			topic := fmt.Sprintf("device/%s/%s/property", d.ProductId, d.Id)
			mqtt.Publish(topic, values)
		}
	}

	//添加计划任务
	if p.Crontab != "" {
		job, err := cron.Crontab(p.Crontab, fn)
		if err != nil {
			return err
		}
		d.jobs = append(d.jobs, job)
	}
	if p.Interval > 0 {
		job, err := cron.Interval(int64(p.Interval), fn)
		if err != nil {
			return err
		}
		d.jobs = append(d.jobs, job)
	}

	devices.Store(d.Id, d)

	return nil
}

func (d *Device) Close() error {
	var err error
	for _, job := range d.jobs {
		e := job.Stop()
		if e != nil {
			err = multierr.Append(err, e)
		}
	}

	devices.Delete(d.Id)

	return err
}

func (d *Device) Poll() (map[string]any, error) {
	if d.product == nil {
		return nil, errors.New("product not exist")
	}
	if d.product.pollers == nil {
		return nil, errors.New("pollers not exist")
	}

	values := map[string]any{}
	for _, p := range d.product.pollers.Pollers {
		buf, err := d.Read(p.Code, p.Address, p.Length)
		if err != nil {
			return nil, err
		}
		//解析
		err = p.Parse(d.product.mappers, buf, values)
		if err != nil {
			return nil, err
		}
	}

	return values, nil
}

func (d *Device) Read(code uint8, offset uint16, length uint16) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(d.Slave)
	buf.WriteByte(code)
	_ = binary.Write(buf, binary.BigEndian, offset)
	_ = binary.Write(buf, binary.BigEndian, length)
	_ = binary.Write(buf, binary.LittleEndian, CRC16(buf.Bytes()))

	//发送
	res, err := d.master.Ask(buf.Bytes(), 7)
	if err != nil {
		return nil, err
	}

	cnt := int(res[2]) //字节数
	ln := 5 + cnt
	//长度不够，继续读
	for len(res) < ln {
		b, e := d.master.Ask(nil, ln-len(res))
		if e != nil {
			return nil, e
		}
		res = append(res, b...)
	}

	return res[3 : len(res)-2], nil //除去包头和crc校验码
}

func (d *Device) Write(code uint8, offset uint16, value any) error {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(d.Slave)
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
	_, err := d.master.Ask(buf.Bytes(), buf.Len()) //写数据时，返回数据一样，长度也一样

	return err
}

func (d *Device) Get(key string) (any, error) {
	if d.product == nil {
		return nil, errors.New("product not exist")
	}
	if d.product.mappers == nil {
		return nil, errors.New("mappers not exist")
	}

	pt, code, addr, size := d.product.mappers.Lookup(key)
	if pt == nil {
		return nil, errors.New("point not exist")
	}

	buf, err := d.Read(code, addr, size)
	if err != nil {
		return nil, err
	}

	return pt.Parse(0, buf)
}

func (d *Device) Set(key string, value any) error {
	if d.product == nil {
		return errors.New("product not exist")
	}
	if d.product.mappers == nil {
		return errors.New("mappers not exist")
	}

	pt, code, addr, _ := d.product.mappers.Lookup(key)
	if pt == nil {
		return errors.New("point not exist")
	}

	buf, err := pt.Encode(value)
	if err != nil {
		return err
	}

	err = d.Write(code, addr, buf)
	if err != nil {
		return err
	}

	return nil
}
