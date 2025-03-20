package internal

import (
	"errors"
	"fmt"
	"github.com/busy-cloud/boat/cron"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/busy-cloud/iot/device"
	"github.com/busy-cloud/iot/product"
	"go.uber.org/multierr"
)

type Station struct {
	Slave uint8 `json:"slave,omitempty"` //从站号
}

type Device struct {
	device.Device `xorm:"extends"`

	LinkerId   string `json:"linker_id,omitempty" xorm:"index"`
	IncomingId string `json:"incoming_id,omitempty" xorm:"index"`

	Station Station `json:"station,omitempty" xorm:"json"`

	master *ModbusMaster

	modbus *Modbus

	jobs []*cron.Job
}

func (d *Device) Open() (err error) {
	//err = json.Unmarshal([]byte(d.Station), &d.station)
	//if err != nil {
	//	return err
	//}

	err, d.modbus = product.LoadConfig[Modbus](d.ProductId, "modbus")
	if err != nil {
		return err
	}

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
	if d.modbus.Crontab != "" {
		job, err := cron.Crontab(d.modbus.Crontab, fn)
		if err != nil {
			return err
		}
		d.jobs = append(d.jobs, job)
	}
	if d.modbus.Interval > 0 {
		job, err := cron.Interval(int64(d.modbus.Interval), fn)
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
	if d.modbus == nil || d.modbus.Pollers == nil {
		return nil, errors.New("pollers not exist")
	}

	values := map[string]any{}
	for _, p := range d.modbus.Pollers {
		buf, err := d.master.Read(d.Station.Slave, p.Code, p.Address, p.Length)
		if err != nil {
			return nil, err
		}
		//解析
		err = p.Parse(d.modbus.Mapper, buf, values)
		if err != nil {
			return nil, err
		}
	}

	return values, nil
}

func (d *Device) Get(key string) (any, error) {
	if d.modbus == nil || d.modbus.Mapper == nil {
		return nil, errors.New("mappers not exist")
	}

	pt, code, addr, size := d.modbus.Mapper.Lookup(key)
	if pt == nil {
		return nil, errors.New("point not exist")
	}

	buf, err := d.master.Read(d.Station.Slave, code, addr, size)
	if err != nil {
		return nil, err
	}

	return pt.Parse(0, buf)
}

func (d *Device) Set(key string, value any) error {
	if d.modbus == nil || d.modbus.Mapper == nil {
		return errors.New("mappers not exist")
	}

	pt, code, addr, _ := d.modbus.Mapper.Lookup(key)
	if pt == nil {
		return errors.New("point not exist")
	}

	buf, err := pt.Encode(value)
	if err != nil {
		return err
	}

	//将读指令变为写指令
	switch code {
	case 1, 2:
		code = 5
	case 3, 4:
		code = 6
	default:
		return fmt.Errorf("invalid code %d", code)
	}

	err = d.master.Write(d.Station.Slave, code, addr, buf)
	if err != nil {
		return err
	}

	return nil
}
