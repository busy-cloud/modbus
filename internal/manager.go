package internal

import (
	"encoding/json"
	"github.com/busy-cloud/boat/lib"
	"github.com/busy-cloud/boat/log"
	"github.com/god-jason/iot-master/product"
	"github.com/god-jason/iot-master/protocol"
)

type Manager struct {
	masters lib.Map[ModbusMaster]
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

func (m *Manager) Model(product_id string, model *product.ProductModel) {
	//models.Store(product_id, model)
	//TODO 解析地址点表
	var cfg ModbusConfig
	configs.Store(product_id, &cfg)

	for _, property := range model.Properties {
		for _, point := range property.Points {
			switch point["register"] {
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
