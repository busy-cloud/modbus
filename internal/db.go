package internal

import "github.com/busy-cloud/boat/db"

func init() {
	db.Register(&Device{}, &ModbusMaster{})
}
