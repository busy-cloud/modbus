package internal

import (
	"github.com/busy-cloud/boat/api"
	"github.com/gin-gonic/gin"
)

func init() {
	api.Register("GET", "modbus-rtu/device/:id/poll", devicePoll)
	api.Register("GET", "modbus-rtu/device/:id/get/:key", deviceGetValue)
	api.Register("POST", "modbus-rtu/device/:id/set/:key", deviceSetValue)
}

func devicePoll(ctx *gin.Context) {
	d := GetDevice(ctx.Param("id"))
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
	d := GetDevice(ctx.Param("id"))
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
	d := GetDevice(ctx.Param("id"))
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
