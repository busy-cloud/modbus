package modbus

import (
	_ "embed"
	"encoding/json"
	"github.com/busy-cloud/boat/log"
	_ "github.com/busy-cloud/connector/boot"
	"github.com/busy-cloud/iot/protocol"
)

//go:embed modbus.json
var manifest string

func init() {

	var p protocol.Protocol
	err := json.Unmarshal([]byte(manifest), &p)
	if err != nil {
		log.Fatal(err)
	}

	protocol.Register(&p)
}
