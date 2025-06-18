package internal

import (
	_ "embed"
	"github.com/busy-cloud/boat/log"
	"github.com/bytedance/sonic"
	"github.com/god-jason/iot-master/protocol"
)

//go:embed protocol.json
var manifest []byte

func init() {
	var p protocol.Protocol
	err := sonic.Unmarshal(manifest, &p)
	if err != nil {
		log.Fatal(err)
	}

	protocol.Create(&p, &Manager{})
}
