package internal

import (
	"cmp"
	"encoding/json"
	"slices"

	"github.com/busy-cloud/boat/lib"
	"github.com/busy-cloud/boat/log"
	"github.com/god-jason/iot-master/product"
	"github.com/god-jason/iot-master/protocol"
	"github.com/spf13/cast"
)

type Manager struct {
	masters lib.Map[ModbusMaster]
}

func (m *Manager) Get(link_id string) protocol.Master {
	return m.masters.Load(link_id)
}

func (m *Manager) Close(link_id string) error {
	master := m.masters.Load(link_id)
	if master != nil {
		return master.Close()
	}
	return nil
}

func (m *Manager) Create(linker, link_id string, options []byte, writer protocol.WriteLinkFunc) (protocol.Master, error) {

	var ops Options
	err := json.Unmarshal(options, &ops)
	if err != nil {
		return nil, err
	}

	var master ModbusMaster
	master.Linker = linker
	master.LinkId = link_id
	master.Options = &ops
	master.writer = writer

	old := m.masters.LoadAndStore(link_id, &master)
	if old != nil {
		_ = old.Close()
	}

	err = master.Open()
	if err != nil {
		return nil, err
	}

	return &master, nil
}

func parseStruct[T any](data map[string]any) (*T, error) {
	var t T
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &t)
	return &t, err
}

func sortPointBit(e *protocol.PointBit, e2 *protocol.PointBit) int {
	return cmp.Compare(e.Address, e2.Address)
}

func sortPointWord(e *protocol.PointWord, e2 *protocol.PointWord) int {
	return cmp.Compare(e.Address, e2.Address)
}

func (m *Manager) Model(product_id string, model *product.ProductModel) {
	//models.Store(product_id, model)

	var cfg ModbusConfig
	configs.Store(product_id, &cfg)

	//解析地址点表
	for _, property := range model.Properties {
		for _, point := range property.Points {
			switch cast.ToInt(point["register"]) {
			case 1:
				p, err := parseStruct[protocol.PointBit](point)
				if err != nil {
					log.Error(err)
				} else {
					cfg.Mapper.Coils = append(cfg.Mapper.Coils, p)
				}
			case 2:
				p, err := parseStruct[protocol.PointBit](point)
				if err != nil {
					log.Error(err)
				} else {
					cfg.Mapper.DiscreteInputs = append(cfg.Mapper.DiscreteInputs, p)
				}
			case 3:
				p, err := parseStruct[protocol.PointWord](point)
				if err != nil {
					log.Error(err)
				} else {
					cfg.Mapper.HoldingRegisters = append(cfg.Mapper.HoldingRegisters, p)
				}
			case 4:
				p, err := parseStruct[protocol.PointWord](point)
				if err != nil {
					log.Error(err)
				} else {
					cfg.Mapper.InputRegisters = append(cfg.Mapper.InputRegisters, p)
				}
			default:
				log.Error("Unknown register type:", point["register"])
			}
		}
	}

	//排序点表
	slices.SortFunc(cfg.Mapper.Coils, sortPointBit)
	slices.SortFunc(cfg.Mapper.DiscreteInputs, sortPointBit)
	slices.SortFunc(cfg.Mapper.HoldingRegisters, sortPointWord)
	slices.SortFunc(cfg.Mapper.InputRegisters, sortPointWord)

	//过滤掉只读点位
	var coils []*protocol.PointBit
	for _, v := range cfg.Mapper.Coils {
		if v.Mode != "w" {
			coils = append(coils, v)
		}
	}
	var discreteInputs []*protocol.PointBit
	for _, v := range cfg.Mapper.DiscreteInputs {
		if v.Mode != "w" {
			discreteInputs = append(discreteInputs, v)
		}
	}
	var holdingRegisters []*protocol.PointWord
	for _, v := range cfg.Mapper.HoldingRegisters {
		if v.Mode != "w" {
			holdingRegisters = append(holdingRegisters, v)
		}
	}
	var inputRegisters []*protocol.PointWord
	for _, V := range cfg.Mapper.InputRegisters {
		if V.Mode != "w" {
			inputRegisters = append(inputRegisters, V)
		}
	}

	//形成轮询器
	if len(coils) > 0 {
		var begin = coils[0]
		var last = begin
		for i := 1; i < len(coils); i++ {
			//出现间隔，就形成一条轮询
			if coils[i].Address > last.Address+1 {
				cfg.Pollers = append(cfg.Pollers, Poller{Code: 1, Address: begin.Address, Length: last.Address - begin.Address + 1})
				begin = coils[i]
				last = begin
			}
			last = coils[i]
		}
		cfg.Pollers = append(cfg.Pollers, Poller{Code: 1, Address: begin.Address, Length: last.Address - begin.Address + 1})
	}
	if len(discreteInputs) > 0 {
		var begin = discreteInputs[0]
		var last = begin
		for i := 1; i < len(discreteInputs); i++ {
			//出现间隔，就形成一条轮询
			if discreteInputs[i].Address > last.Address+1 {
				cfg.Pollers = append(cfg.Pollers, Poller{Code: 2, Address: begin.Address, Length: last.Address - begin.Address + 1})
				begin = discreteInputs[i]
				last = begin
			}
			last = discreteInputs[i]
		}
		cfg.Pollers = append(cfg.Pollers, Poller{Code: 2, Address: begin.Address, Length: last.Address - begin.Address + 1})
	}
	if len(holdingRegisters) > 0 {
		var begin = holdingRegisters[0]
		var last = begin
		for i := 1; i < len(holdingRegisters); i++ {
			//出现间隔，就形成一条轮询
			if holdingRegisters[i].Address > last.Address+1 {
				cfg.Pollers = append(cfg.Pollers, Poller{Code: 3, Address: begin.Address, Length: last.Address - begin.Address + uint16(last.Size())})
				begin = holdingRegisters[i]
				last = begin
			}
			last = holdingRegisters[i]
		}
		cfg.Pollers = append(cfg.Pollers, Poller{Code: 3, Address: begin.Address, Length: last.Address - begin.Address + uint16(last.Size())})
	}
	if len(inputRegisters) > 0 {
		var begin = inputRegisters[0]
		var last = begin
		for i := 1; i < len(inputRegisters); i++ {
			//出现间隔，就形成一条轮询
			if inputRegisters[i].Address > last.Address+1 {
				cfg.Pollers = append(cfg.Pollers, Poller{Code: 4, Address: begin.Address, Length: last.Address - begin.Address + uint16(last.Size())})
				begin = inputRegisters[i]
				last = begin
			}
			last = inputRegisters[i]
		}
		cfg.Pollers = append(cfg.Pollers, Poller{Code: 4, Address: begin.Address, Length: last.Address - begin.Address + uint16(last.Size())})
	}

}
