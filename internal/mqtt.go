package internal

import (
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	paho "github.com/eclipse/paho.mqtt.golang"
	"strings"
)

const protocol = "modbus"

func Startup() error {

	//订阅数据
	mqtt.Client.SubscribeMultiple(map[string]byte{
		protocol + "/+/+/up": 0,
		protocol + "/+/up":   0,
	}, func(client paho.Client, message paho.Message) {
		ss := strings.Split(message.Topic(), "/")
		linker := ss[1]
		incoming := ss[2]
		gateway, err := EnsureGateway(linker, incoming)
		if err != nil {
			log.Error("gateway err:", err)
			return
		}
		gateway.onData(message.Payload())
	})

	//连接打开，加载设备
	mqtt.Client.SubscribeMultiple(map[string]byte{
		protocol + "/+/+/open": 0,
		protocol + "/+/open":   0,
	}, func(client paho.Client, message paho.Message) {
		ss := strings.Split(message.Topic(), "/")
		linker := ss[1]
		incoming := ss[2]
		gateway, err := EnsureGateway(linker, incoming)
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
	mqtt.Client.SubscribeMultiple(map[string]byte{
		protocol + "/+/+/close": 0,
		protocol + "/+/close":   0,
	}, func(client paho.Client, message paho.Message) {
		ss := strings.Split(message.Topic(), "/")
		linker := ss[1]
		incoming := ss[2]
		gateway, err := EnsureGateway(linker, incoming)
		if err != nil {
			log.Error("gateway err:", err)
			return
		}
		if gateway != nil {
			//gateway.Close()
		}
	})

	////添加设备
	//mqtt.Subscribe("link/+/+/attach/+", func(topic string, payload []byte) {
	//	ss := strings.Split(topic, "/")
	//	id := ss[2]
	//	id2 := ss[4]
	//	gateway := GetGateway(id)
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
	//	gateway := GetGateway(id)
	//	if gateway != nil {
	//		gateway.UnLoadDevice(id2)
	//	}
	//})

	return nil
}
