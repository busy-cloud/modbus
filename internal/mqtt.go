package internal

import (
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"strings"
)

const protocol = "modbus"

func Startup() error {

	//订阅数据
	mqtt.Subscribe(protocol+"/+/+/up", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		linker := ss[1]
		incoming := ss[2]
		master, err := EnsureMaster(linker, incoming)
		if err != nil {
			log.Error("master err:", err)
			return
		}
		master.onData(payload)
	})
	mqtt.Subscribe(protocol+"/+/up", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		linker := ss[1]
		master, err := EnsureMaster(linker, "")
		if err != nil {
			log.Error("master err:", err)
			return
		}
		master.onData(payload)
	})

	//连接打开，加载设备
	mqtt.Subscribe(protocol+"/+/+/open", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		linker := ss[1]
		incoming := ss[2]
		master, err := EnsureMaster(linker, incoming)
		if err != nil {
			log.Error("master err:", err)
			return
		}
		err = master.Open()
		//log.Println("master open", master.Id)
		if err != nil {
			log.Println(err)
		}
	})
	mqtt.Subscribe(protocol+"/+/open", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		linker := ss[1]
		master, err := EnsureMaster(linker, "")
		if err != nil {
			log.Error("master err:", err)
			return
		}
		err = master.Open()
		//log.Println("master open", master.Id)
		if err != nil {
			log.Println(err)
		}
	})

	//关闭连接
	mqtt.Subscribe(protocol+"/+/+/close", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		linker := ss[1]
		incoming := ss[2]
		master := GetMasterLinkerAndIncoming(linker, incoming)
		if master != nil {
			_ = master.Close()
		}
	})
	mqtt.Subscribe(protocol+"/+/close", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		linker := ss[1]
		master := GetMasterLinkerAndIncoming(linker, "")
		if master != nil {
			_ = master.Close()
		}
	})

	////添加设备
	//mqtt.Subscribe("link/+/+/attach/+", func(topic string, payload []byte) {
	//	ss := strings.Split(topic, "/")
	//	id := ss[2]
	//	id2 := ss[4]
	//	master := GetMaster(id)
	//	if master != nil {
	//		err := master.LoadDevice(id2)
	//		if err != nil {
	//			log.Error("master err:", err)
	//		}
	//	}
	//})
	//
	////删除设备
	//mqtt.Subscribe("link/+/+/detach/+", func(topic string, payload []byte) {
	//	ss := strings.Split(topic, "/")
	//	id := ss[2]
	//	id2 := ss[4]
	//	master := GetMaster(id)
	//	if master != nil {
	//		master.UnLoadDevice(id2)
	//	}
	//})

	return nil
}
