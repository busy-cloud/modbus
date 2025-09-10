package internal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/god-jason/iot-master/calc"
)

type Station struct {
	Slave uint8 `json:"slave,omitempty"` //从站号
}

type Device struct {
	Id        string  `json:"id,omitempty" xorm:"pk"`
	ProductId string  `json:"product_id,omitempty"`
	Station   Station `json:"station,omitempty" xorm:"json"`

	master *ModbusMaster
}

func (d *Device) Poll() (map[string]any, error) {
	config := configs.Load(d.ProductId)

	if config == nil || config.Pollers == nil {
		return nil, errors.New("pollers not exist")
	}

	values := map[string]any{}
	for _, p := range config.Pollers {
		buf, err := d.master.Read(d.Station.Slave, p.Code, p.Address, p.Length)
		if err != nil {
			return nil, err
		}
		//解析
		err = p.Parse(&config.Mapper, buf, values)
		if err != nil {
			return nil, err
		}
	}

	return values, nil
}

func (d *Device) Get(key string) (any, error) {
	config := configs.Load(d.ProductId)

	if config == nil {
		return nil, errors.New("model not exist")
	}

	pt, code, addr, size := config.Mapper.Lookup(key)
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
	config := configs.Load(d.ProductId)

	if config == nil {
		return errors.New("model not exist")
	}

	pt, code, addr, _ := config.Mapper.Lookup(key)
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

func (d *Device) Action(name string, args map[string]any) error {

	config := configs.Load(d.ProductId)
	if config == nil || config.Actions == nil {
		return errors.New("actions not exist")
	}

	var action *Action
	for _, a := range config.Actions {
		if a.Name == name {
			action = a
			break
		}
	}
	if action == nil {
		return errors.New("action not exist")
	}

	for _, o := range action.Operators {
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
