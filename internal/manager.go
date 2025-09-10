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

	//形成轮询器
	if len(cfg.Mapper.Coils) > 0 {
		var begin = cfg.Mapper.Coils[0]
		var last = begin
		for i := 1; i < len(cfg.Mapper.Coils); i++ {
			//出现间隔，就形成一条轮询
			if cfg.Mapper.Coils[i].Address > last.Address+1 {
				cfg.Pollers = append(cfg.Pollers, Poller{Code: 1, Address: begin.Address, Length: last.Address - begin.Address + 1})
				begin = cfg.Mapper.Coils[i]
				last = begin
			}
			last = cfg.Mapper.Coils[i]
		}
		cfg.Pollers = append(cfg.Pollers, Poller{Code: 1, Address: begin.Address, Length: last.Address - begin.Address + 1})
	}
	if len(cfg.Mapper.DiscreteInputs) > 0 {
		var begin = cfg.Mapper.DiscreteInputs[0]
		var last = begin
		for i := 1; i < len(cfg.Mapper.DiscreteInputs); i++ {
			//出现间隔，就形成一条轮询
			if cfg.Mapper.DiscreteInputs[i].Address > last.Address+1 {
				cfg.Pollers = append(cfg.Pollers, Poller{Code: 2, Address: begin.Address, Length: last.Address - begin.Address + 1})
				begin = cfg.Mapper.DiscreteInputs[i]
				last = begin
			}
			last = cfg.Mapper.DiscreteInputs[i]
		}
		cfg.Pollers = append(cfg.Pollers, Poller{Code: 2, Address: begin.Address, Length: last.Address - begin.Address + 1})
	}
	if len(cfg.Mapper.HoldingRegisters) > 0 {
		var begin = cfg.Mapper.HoldingRegisters[0]
		var last = begin
		for i := 1; i < len(cfg.Mapper.HoldingRegisters); i++ {
			//出现间隔，就形成一条轮询
			if cfg.Mapper.HoldingRegisters[i].Address > last.Address+1 {
				cfg.Pollers = append(cfg.Pollers, Poller{Code: 3, Address: begin.Address, Length: last.Address - begin.Address + uint16(last.Size())})
				begin = cfg.Mapper.HoldingRegisters[i]
				last = begin
			}
			last = cfg.Mapper.HoldingRegisters[i]
		}
		cfg.Pollers = append(cfg.Pollers, Poller{Code: 3, Address: begin.Address, Length: last.Address - begin.Address + uint16(last.Size())})
	}
	if len(cfg.Mapper.InputRegisters) > 0 {
		var begin = cfg.Mapper.InputRegisters[0]
		var last = begin
		for i := 1; i < len(cfg.Mapper.InputRegisters); i++ {
			//出现间隔，就形成一条轮询
			if cfg.Mapper.InputRegisters[i].Address > last.Address+1 {
				cfg.Pollers = append(cfg.Pollers, Poller{Code: 4, Address: begin.Address, Length: last.Address - begin.Address + uint16(last.Size())})
				begin = cfg.Mapper.InputRegisters[i]
				last = begin
			}
			last = cfg.Mapper.InputRegisters[i]
		}
		cfg.Pollers = append(cfg.Pollers, Poller{Code: 4, Address: begin.Address, Length: last.Address - begin.Address + uint16(last.Size())})
	}

}
