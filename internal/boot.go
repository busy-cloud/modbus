package internal

import (
	"github.com/busy-cloud/boat/boot"
)

func init() {
	boot.Register("modbus", &boot.Task{
		Startup:  Startup, //启动
		Shutdown: nil,
		Depends:  []string{"web", "log", "database", "mqtt", "connector"},
	})
}
