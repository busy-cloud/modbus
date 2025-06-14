package internal

import (
	_ "embed"
	"encoding/json"
	"github.com/busy-cloud/boat/log"
	"github.com/god-jason/iot-master/protocol"
)

//go:embed protocol.json
var manifest []byte

func init() {
	var p protocol.Protocol
	err := json.Unmarshal(manifest, &p)
	if err != nil {
		log.Fatal(err)
	}

	protocol.Create(&p, &Manager{})
}
