package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/busy-cloud/boat/boot"
	_ "github.com/busy-cloud/boat/broker"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/web"
	_ "github.com/busy-cloud/modbus"
	_ "github.com/busy-cloud/tcp-client"
	_ "github.com/god-jason/iot-master"
	"github.com/spf13/viper"
)

func main() {
	//viper.SetConfigName("modbus")
	//viper.SetConfigType("yaml")
	viper.SetConfigFile("modbus.yaml")

	//注册系统信号
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs

		//关闭web，出发
		_ = web.Shutdown()
	}()

	//安全退出
	defer boot.Shutdown()

	//正式启用
	err := boot.Startup()
	if err != nil {
		log.Fatal(err)
		return
	}

	//静态目录
	web.StaticDir("www", "", "", "index.html")

	err = web.Serve()
	if err != nil {
		log.Fatal(err)
		return
	}
}
