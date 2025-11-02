package internal

import (
	"github.com/god-jason/iot-master/protocol"
)

type Mapper struct {
	Coils            []*protocol.PointBit  `json:"coils,omitempty"`
	DiscreteInputs   []*protocol.PointBit  `json:"discrete_inputs,omitempty"`
	HoldingRegisters []*protocol.PointWord `json:"holding_registers,omitempty"`
	InputRegisters   []*protocol.PointWord `json:"input_registers,omitempty"`
}

func (p *Mapper) Lookup(name string) (pt protocol.Point, code uint8, address uint16, size uint16) {
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

func (p *Mapper) LookupRead(name string) (pt protocol.Point, code uint8, address uint16, size uint16) {
	for _, m := range p.Coils {
		if m.Name == name && m.Mode != "w" {
			return m, 1, m.Address, 1
		}
	}

	for _, m := range p.DiscreteInputs {
		if m.Name == name && m.Mode != "w" {
			return m, 2, m.Address, 1
		}
	}

	for _, m := range p.HoldingRegisters {
		if m.Name == name && m.Mode != "w" {
			return m, 3, m.Address, uint16(m.Size())
		}
	}

	for _, m := range p.InputRegisters {
		if m.Name == name && m.Mode != "w" {
			return m, 4, m.Address, uint16(m.Size())
		}
	}
	return
}

func (p *Mapper) LookupWrite(name string) (pt protocol.Point, code uint8, address uint16, size uint16) {
	for _, m := range p.Coils {
		if m.Name == name && m.Mode != "r" {
			return m, 1, m.Address, 1
		}
	}

	for _, m := range p.DiscreteInputs {
		if m.Name == name && m.Mode != "r" {
			return m, 2, m.Address, 1
		}
	}

	for _, m := range p.HoldingRegisters {
		if m.Name == name && m.Mode != "r" {
			return m, 3, m.Address, uint16(m.Size())
		}
	}

	for _, m := range p.InputRegisters {
		if m.Name == name && m.Mode != "r" {
			return m, 4, m.Address, uint16(m.Size())
		}
	}
	return
}
