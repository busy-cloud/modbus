package apis

import (
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
	api.Register("POST", "modbus/device/:id", curd.ApiUpdate[internal.Device]("id", "name", "description", "product_id", "disabled", "master_id", "slave"))
	api.Register("GET", "modbus/device/:id/delete", curd.ApiDelete[internal.Device]())
	api.Register("GET", "modbus/device/:id/enable", curd.ApiDisable[internal.Device](false))
	api.Register("GET", "modbus/device/:id/disable", curd.ApiDisable[internal.Device](true))
	api.Register("GET", "modbus/device/:id/poll", devicePoll)
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
