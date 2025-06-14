package main

import (
	"github.com/busy-cloud/boat/boot"
	"github.com/busy-cloud/boat/log"
	_ "github.com/busy-cloud/modbus/internal" //引入主程序
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	viper.SetConfigName("modbus")

	//注册系统信号
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs

		//关闭web，出发
		os.Exit(0)
	}()

	//安全退出
	defer boot.Shutdown()

	//正式启用
	err := boot.Startup()
	if err != nil {
		log.Fatal(err)
		return
	}

	select {}
}
