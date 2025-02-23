package internal

import (
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"strings"
)

func Startup() error {

	//订阅数据
	mqtt.Subscribe("link/base/+/up", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		id := ss[2]
		gateway, err := EnsureGateway(id)
		if err != nil {
			log.Error("gateway err:", err)
			return
		}
		gateway.onData(payload)
	})

	//连接打开，加载设备
	mqtt.Subscribe("link/base/+/open", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		id := ss[2]
		gateway, err := EnsureGateway(id)
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
	mqtt.Subscribe("link/base/+/close", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		id := ss[2]
		gateway := GetGateway(id)
		if gateway != nil {
			//gateway.Close()
		}
	})

	//添加设备
	mqtt.Subscribe("link/base/+/attach/+", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		id := ss[2]
		id2 := ss[4]
		gateway := GetGateway(id)
		if gateway != nil {
			err := gateway.LoadDevice(id2)
			if err != nil {
				log.Error("gateway err:", err)
			}
		}
	})

	//删除设备
	mqtt.Subscribe("link/base/+/detach/+", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		id := ss[2]
		id2 := ss[4]
		gateway := GetGateway(id)
		if gateway != nil {
			gateway.UnLoadDevice(id2)
		}
	})

	return nil
}
