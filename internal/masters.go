package internal

import (
	"github.com/busy-cloud/boat/lib"
)

var masters lib.Map[ModbusMaster]

func CombineId(linker, incoming string) string {
	if incoming == "" {
		return linker
	}
	return linker + "_" + incoming
}

func GetMaster(linker, incoming string) *ModbusMaster {
	id := CombineId(linker, incoming)
	return masters.Load(id)
}

func CreateMaster(linker, incoming string, options *Options) (*ModbusMaster, error) {
	id := CombineId(linker, incoming)

	//从数据库加载
	var master ModbusMaster
	master.Linker = linker
	master.LinkId = incoming
	master.Options = options

	old := masters.LoadAndStore(id, &master)
	if old != nil {
		_ = old.Close()
	}

	err := master.Open()
	if err != nil {
		return nil, err
	}

	return &master, nil
}
