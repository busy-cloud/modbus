package internal

import "github.com/busy-cloud/iot/product"

type Mappers struct {
	Coils            []*product.PointBit  `json:"coils,omitempty"`
	DiscreteInputs   []*product.PointBit  `json:"discrete_inputs,omitempty"`
	HoldingRegisters []*product.PointWord `json:"holding_registers,omitempty"`
	InputRegisters   []*product.PointWord `json:"input_registers,omitempty"`
}

func (p *Mappers) Lookup(name string) (pt product.Point, code uint8, address uint16, size uint16) {
	for _, m := range p.Coils {
		if m.Name == name {
			return m, 1, m.Address, 1
		}
	}

	for _, m := range p.DiscreteInputs {
		if m.Name == name {
			return m, 2, m.Address, 1
		}
	}

	for _, m := range p.HoldingRegisters {
		if m.Name == name {
			return m, 3, m.Address, uint16(m.Size())
		}
	}

	for _, m := range p.InputRegisters {
		if m.Name == name {
			return m, 4, m.Address, uint16(m.Size())
		}
	}
	return
}
