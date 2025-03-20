package modbus

import (
	_ "embed"
	"encoding/json"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/iot/protocol"
	_ "github.com/busy-cloud/modbus/internal"
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
