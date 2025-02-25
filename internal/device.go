package internal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/iot/types"
	"github.com/spf13/cast"
)

func init() {
	db.Register(&Device{})
}

type Device struct {
	types.Device `xorm:"extends"`
	Slave        uint8 `json:"slave,omitempty"` //从站号

	gateway *ModbusMaster
}

func (d *Device) Read(code uint8, offset uint16, length uint16) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(d.Slave)
	buf.WriteByte(code)
	_ = binary.Write(buf, binary.BigEndian, offset)
	_ = binary.Write(buf, binary.BigEndian, length)
	_ = binary.Write(buf, binary.LittleEndian, CRC16(buf.Bytes()))

	//发送
	err := d.gateway.Write(buf.Bytes())
	if err != nil {
		return nil, err
	}

	res, err := d.gateway.ReadAtLeast(7)
	if err != nil {
		return nil, err
	}

	cnt := int(res[2]) //字节数
	ln := 5 + cnt
	//长度不够，继续读
	for len(res) < ln {
		b, e := d.gateway.ReadAtLeast(ln - len(res))
		if e != nil {
			return nil, e
		}
		res = append(res, b...)
	}

	return res[3 : len(res)-2], nil //除去包头和crc校验码
}

func (d *Device) Write(code uint8, offset uint16, value any) error {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(d.Slave)
	buf.WriteByte(code)
	_ = binary.Write(buf, binary.BigEndian, offset)
	switch code {
	case 5: //单个线圈
		if cast.ToBool(value) {
			buf.WriteByte(0xff)
			buf.WriteByte(0xff)
		} else {
			buf.WriteByte(0x00)
			buf.WriteByte(0xff)
		}
	//case 15: //多个线圈
	case 6: //单个寄存器
		_ = binary.Write(buf, binary.BigEndian, cast.ToUint16(value))
	//case 16: //多个寄存器
	default:
		return fmt.Errorf("invalid code: %d", code)
	}

	_ = binary.Write(buf, binary.LittleEndian, CRC16(buf.Bytes()))

	//发送
	err := d.gateway.Write(buf.Bytes())
	if err != nil {
		return err
	}

	_, err = d.gateway.ReadAtLeast(buf.Len()) //写数据时，返回数据一样，长度也一样

	return err
}
