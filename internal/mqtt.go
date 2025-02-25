package internal

import (
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"strings"
)

const protocol = "modbus"

func Startup() error {

	scheduler.Start()

	//订阅数据
	mqtt.Subscribe(protocol+"/+/+/up", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		linker := ss[1]
		incoming := ss[2]
		gateway, err := EnsureMaster(linker, incoming)
		if err != nil {
			log.Error("gateway err:", err)
			return
		}
		gateway.onData(payload)
	})
	mqtt.Subscribe(protocol+"/+/up", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		linker := ss[1]
		gateway, err := EnsureMaster(linker, "")
		if err != nil {
			log.Error("gateway err:", err)
			return
		}
		gateway.onData(payload)
	})

	//连接打开，加载设备
	mqtt.Subscribe(protocol+"/+/+/open", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		linker := ss[1]
		incoming := ss[2]
		gateway, err := EnsureMaster(linker, incoming)
		if err != nil {
			log.Error("gateway err:", err)
			return
		}
		err = gateway.Open()
		//log.Println("gateway open", gateway.Id)
		if err != nil {
			log.Println(err)
		}
	})
	mqtt.Subscribe(protocol+"/+/open", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		linker := ss[1]
		gateway, err := EnsureMaster(linker, "")
		if err != nil {
			log.Error("gateway err:", err)
			return
		}
		err = gateway.Open()
		//log.Println("gateway open", gateway.Id)
		if err != nil {
			log.Println(err)
		}
	})

	//关闭连接
	mqtt.Subscribe(protocol+"/+/+/close", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		linker := ss[1]
		incoming := ss[2]
		gateway := GetMaster(linker, incoming)
		if gateway != nil {
			_ = gateway.Close()
		}
	})
	mqtt.Subscribe(protocol+"/+/close", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		linker := ss[1]
		gateway := GetMaster(linker, "")
		if gateway != nil {
			_ = gateway.Close()
		}
	})

	////添加设备
	//mqtt.Subscribe("link/+/+/attach/+", func(topic string, payload []byte) {
	//	ss := strings.Split(topic, "/")
	//	id := ss[2]
	//	id2 := ss[4]
	//	gateway := GetMaster(id)
	//	if gateway != nil {
	//		err := gateway.LoadDevice(id2)
	//		if err != nil {
	//			log.Error("gateway err:", err)
	//		}
	//	}
	//})
	//
	////删除设备
	//mqtt.Subscribe("link/+/+/detach/+", func(topic string, payload []byte) {
	//	ss := strings.Split(topic, "/")
	//	id := ss[2]
	//	id2 := ss[4]
	//	gateway := GetMaster(id)
	//	if gateway != nil {
	//		gateway.UnLoadDevice(id2)
	//	}
	//})

	return nil
}
