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
		master := GetMaster(linker, incoming)
		if master != nil {
			master.onData(payload)
		}
		master.onData(payload)
	})

	mqtt.Subscribe(protocol+"/+/up", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		linker := ss[1]
		master := GetMaster(linker, "")
		if master != nil {
			master.onData(payload)
		}
	})

	//连接打开，加载设备
	mqtt.SubscribeStruct[Options](protocol+"/+/+/open", func(topic string, options *Options) {
		ss := strings.Split(topic, "/")
		linker := ss[1]
		incoming := ss[2]
		_, err := CreateMaster(linker, incoming, options)
		if err != nil {
			log.Error("master err:", err)
			return
		}
	})

	mqtt.SubscribeStruct(protocol+"/+/open", func(topic string, options *Options) {
		ss := strings.Split(topic, "/")
		linker := ss[1]
		_, err := CreateMaster(linker, "", options)
		if err != nil {
			log.Error("master err:", err)
			return
		}
	})

	//关闭连接
	mqtt.Subscribe(protocol+"/+/+/close", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		linker := ss[1]
		incoming := ss[2]
		master := GetMaster(linker, incoming)
		if master != nil {
			_ = master.Close()
		}
	})

	mqtt.Subscribe(protocol+"/+/close", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		linker := ss[1]
		master := GetMaster(linker, "")
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

func WriteTo(linker, incoming string, data []byte) error {
	topic := "link/" + linker
	if incoming != "" {
		topic += "/" + incoming
	}
	topic += "/down"
	tkn := mqtt.Publish(topic, data)
	tkn.Wait()
	return tkn.Error()
}
