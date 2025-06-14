package internal

import (
	"encoding/json"
	"github.com/busy-cloud/boat/lib"
	"github.com/god-jason/iot-master/protocol"
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
