package internal

import (
	"errors"
	"fmt"
	"github.com/busy-cloud/boat/cron"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/busy-cloud/iot/types"
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
		buf, err := d.master.Read(d.Slave, p.Code, p.Address, p.Length)
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

	buf, err := d.master.Read(d.Slave, code, addr, size)
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

	err = d.master.Write(d.Slave, code, addr, buf)
	if err != nil {
		return err
	}

	return nil
}
