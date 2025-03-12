package boot

import (
	"github.com/busy-cloud/boat/boot"
	"github.com/busy-cloud/modbus/internal"
)

func init() {
	boot.Register("modbus", &boot.Task{
		Startup:  internal.Startup, //启动
		Shutdown: nil,
		Depends:  []string{"web", "log", "database", "mqtt"},
	})
}
