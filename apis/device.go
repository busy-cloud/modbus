package apis

import (
	"encoding/hex"
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
	"github.com/busy-cloud/modbus/internal"
	"github.com/gin-gonic/gin"
)

func init() {
	api.Register("POST", "modbus/device/count", curd.ApiCount[internal.Device]())
	api.Register("POST", "modbus/device/search", curd.ApiSearch[internal.Device]())
	api.Register("GET", "modbus/device/list", curd.ApiList[internal.Device]())
	api.Register("POST", "modbus/device/create", curd.ApiCreate[internal.Device]())
	api.Register("GET", "modbus/device/:id", curd.ApiGet[internal.Device]())
	api.Register("POST", "modbus/device/:id", curd.ApiUpdate[internal.Device]("id", "name", "description", "product_id", "disabled", "linker_id", "incoming_id", "slave"))
	api.Register("GET", "modbus/device/:id/delete", curd.ApiDelete[internal.Device]())
	api.Register("GET", "modbus/device/:id/enable", curd.ApiDisable[internal.Device](false))
	api.Register("GET", "modbus/device/:id/disable", curd.ApiDisable[internal.Device](true))
	api.Register("GET", "modbus/device/:id/poll", devicePoll)
	api.Register("GET", "modbus/device/:id/read", deviceRead)
	api.Register("POST", "modbus/device/:id/write", deviceWrite)
	api.Register("GET", "modbus/device/:id/get/:key", deviceGetValue)
	api.Register("POST", "modbus/device/:id/set/:key", deviceSetValue)
}

func devicePoll(ctx *gin.Context) {
	d := internal.GetDevice(ctx.Param("id"))
	if d == nil {
		api.Fail(ctx, "device not found")
		return
	}

	values, err := d.Poll()
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, values)
}

type deviceReadBody struct {
	Code    uint8  `json:"code"`
	Address uint16 `json:"address"`
	Length  uint16 `json:"length"`
}

func deviceRead(ctx *gin.Context) {
	d := internal.GetDevice(ctx.Param("id"))
	if d == nil {
		api.Fail(ctx, "device not found")
		return
	}

	var body deviceReadBody
	if err := ctx.BindQuery(&body); err != nil {
		api.Error(ctx, err)
		return
	}

	buf, err := d.Read(body.Code, body.Address, body.Length)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, hex.EncodeToString(buf))
}

type deviceWriteBody struct {
	Code    uint8  `json:"code"`
	Address uint16 `json:"address"`
	Value   string `json:"value"`
}

func deviceWrite(ctx *gin.Context) {
	d := internal.GetDevice(ctx.Param("id"))
	if d == nil {
		api.Fail(ctx, "device not found")
		return
	}

	var body deviceWriteBody
	if err := ctx.ShouldBind(&body); err != nil {
		api.Error(ctx, err)
		return
	}

	buf, err := hex.DecodeString(body.Value)

	err = d.Write(body.Code, body.Address, buf)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, nil)
}

func deviceGetValue(ctx *gin.Context) {
	d := internal.GetDevice(ctx.Param("id"))
	if d == nil {
		api.Fail(ctx, "device not found")
		return
	}

	value, err := d.Get(ctx.Param("key"))
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, value)
}

type deviceSetBody struct {
	Value any `json:"value"`
}

func deviceSetValue(ctx *gin.Context) {
	d := internal.GetDevice(ctx.Param("id"))
	if d == nil {
		api.Fail(ctx, "device not found")
		return
	}

	var body deviceSetBody
	if err := ctx.ShouldBind(&body); err != nil {
		api.Error(ctx, err)
		return
	}

	err := d.Set(ctx.Param("key"), body.Value)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, nil)
}
