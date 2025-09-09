package internal

import (
	_ "embed"
	"encoding/json"
	"github.com/busy-cloud/boat/boot"
	"github.com/god-jason/iot-master/protocol"
)

func init() {
	boot.Register("modbus", &boot.Task{
		Startup:  Startup,
		Shutdown: nil,
		Depends:  []string{"log", "mqtt", "iot"},
	})
}

//go:embed protocol.json
var manifest []byte

func Startup() error {

	var p protocol.Protocol
	err := json.Unmarshal(manifest, &p)
	if err != nil {
		return err
	}

	protocol.Create(&p, &Manager{})

	return nil
}
