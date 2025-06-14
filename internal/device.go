package internal

import (
	"context"
	"errors"
	"fmt"
	"github.com/busy-cloud/boat/db"
	"github.com/god-jason/iot-master/calc"
	"github.com/god-jason/iot-master/device"
	"github.com/god-jason/iot-master/product"
	"time"
)

func init() {
	db.Register(&Device{})
}

type Station struct {
	Slave uint8 `json:"slave,omitempty"` //从站号
}

type Device struct {
	device.Device `xorm:"extends"`

	Station Station `json:"station,omitempty" xorm:"json"`

	master *ModbusMaster

	config *ModbusConfig
}

func (d *Device) Open() (err error) {
	//err = json.Unmarshal([]byte(d.Station), &d.station)
	//if err != nil {
	//	return err
	//}

	d.config, err = product.LoadConfig[ModbusConfig](d.ProductId, "modbus")
	if err != nil {
		return err
	}

	devices.Store(d.Id, d)

	return nil
}

func (d *Device) Close() error {
	var err error

	devices.Delete(d.Id)

	return err
}

func (d *Device) Poll() (map[string]any, error) {
	if d.config == nil || d.config.Pollers == nil {
		return nil, errors.New("pollers not exist")
	}

	values := map[string]any{}
	for _, p := range d.config.Pollers {
		buf, err := d.master.Read(d.Station.Slave, p.Code, p.Address, p.Length)
		if err != nil {
			return nil, err
		}
		//解析
		err = p.Parse(d.config.Mapper, buf, values)
		if err != nil {
			return nil, err
		}
	}

	return values, nil
}

func (d *Device) Get(key string) (any, error) {
	if d.config == nil || d.config.Mapper == nil {
		return nil, errors.New("mappers not exist")
	}

	pt, code, addr, size := d.config.Mapper.Lookup(key)
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
	if d.config == nil || d.config.Mapper == nil {
		return errors.New("mappers not exist")
	}

	pt, code, addr, _ := d.config.Mapper.Lookup(key)
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

func (d *Device) Action(operators []*Operator, args map[string]any) error {
	for _, o := range operators {

		expr, err := calc.Compile(o.Value)
		if err != nil {
			return err
		}
		val, err := expr.EvalFloat64(context.Background(), args)
		if err != nil {
			return err
		}

		if o.Delay > 0 {
			time.Sleep(time.Second * time.Duration(o.Delay))
		}

		err = d.Set(o.Name, val)
		if err != nil {
			return err
		}
	}

	return nil
}
