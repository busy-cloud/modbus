package internal

import "github.com/busy-cloud/boat/lib"

var devices lib.Map[Device]

func GetDevice(id string) *Device {
	return devices.Load(id)
}
