package internal

import (
	"errors"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/lib"
)

var masters lib.Map[ModbusMaster]
var mastersByLinkerAndIncoming lib.Map[ModbusMaster]

func CombineId(linker, incoming string) string {
	if incoming == "" {
		return linker
	}
	return linker + "_" + incoming
}

func GetMaster(id string) *ModbusMaster {
	return masters.Load(id)
}

func LoadMaster(id string) error {
	//从数据库加载
	var master ModbusMaster
	has, err := db.Engine().ID(id).Get(&master)
	if err != nil {
		return err
	}
	if !has {
		return errors.New("master not found")
	}

	last := masters.LoadAndStore(master.Id, &master)
	mastersByLinkerAndIncoming.Store(master.ID(), &master)
	if last != nil {
		_ = last.Close()
	}

	err = master.Open()
	if err != nil {
		return err
	}

	return nil
}

func UnloadMaster(id string) error {
	m := masters.LoadAndDelete(id)
	if m != nil {
		mastersByLinkerAndIncoming.Delete(m.ID())
		return m.Close()
	}
	return nil
}

func GetMasterLinkerAndIncoming(linker, incoming string) *ModbusMaster {
	return mastersByLinkerAndIncoming.Load(CombineId(linker, incoming))
}

// EnsureMaster 自动加载网关
func EnsureMaster(linker, incoming string) (*ModbusMaster, error) {
	id := CombineId(linker, incoming)
	m := mastersByLinkerAndIncoming.Load(id)
	if m != nil {
		return m, nil
	}

	//从数据库加载
	var master ModbusMaster
	has, err := db.Engine().Where("linker_id=? AND incoming_id=?", linker, incoming).Get(&master)
	if err != nil {
		return nil, err
	}
	if !has {
		//return nil, errors.New("master not found")
		master.Id = id
		_, err = db.Engine().InsertOne(&master)
		if err != nil {
			return nil, err
		}
	}

	masters.Store(master.Id, &master)
	mastersByLinkerAndIncoming.Store(id, &master)

	err = master.Open()
	if err != nil {
		return nil, err
	}

	return &master, nil
}
