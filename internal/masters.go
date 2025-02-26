package internal

import (
	"errors"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/lib"
)

var masters lib.Map[ModbusMaster]
var mastersByLinkerAndIncoming lib.Map[ModbusMaster]

func GetMaster(id string) *ModbusMaster {
	return masters.Load(id)
}

func LoadMaster(id string) (*ModbusMaster, error) {
	//从数据库加载
	var master ModbusMaster
	has, err := db.Engine.ID(id).Get(&master)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("master not found")
	}

	masters.Store(master.Id, &master)
	LinkerAndIncoming := master.LinkerId + "_" + master.IncomingId
	mastersByLinkerAndIncoming.Store(LinkerAndIncoming, &master)

	return &master, nil
}

func Unload(id string) {

}

func GetMasterLinkerAndIncoming(linker, incoming string) *ModbusMaster {
	LinkerAndIncoming := linker + "_" + incoming
	return mastersByLinkerAndIncoming.Load(LinkerAndIncoming)
}

// EnsureMaster 自动加载网关
func EnsureMaster(linker, incoming string) (*ModbusMaster, error) {
	LinkerAndIncoming := linker + "_" + incoming
	m := mastersByLinkerAndIncoming.Load(LinkerAndIncoming)
	if m != nil {
		return m, nil
	}

	//从数据库加载
	var master ModbusMaster
	has, err := db.Engine.Where("linker_id=? AND incoming_id=?", linker, incoming).Get(&master)
	if err != nil {
		return nil, err
	}
	if !has {
		//return nil, errors.New("master not found")
		master.Id = LinkerAndIncoming
		_, err = db.Engine.InsertOne(&master)
		if err != nil {
			return nil, err
		}
	}

	masters.Store(master.Id, &master)
	mastersByLinkerAndIncoming.Store(LinkerAndIncoming, &master)

	return &master, nil
}
