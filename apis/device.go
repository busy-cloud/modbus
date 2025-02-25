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
	api.Register("GET", "modbus/device/:id", curd.ParseParamStringId, curd.ApiGet[internal.Device]())
	api.Register("POST", "modbus/device/:id", curd.ParseParamStringId, curd.ApiUpdate[internal.Device]("id", "name", "description", "product_id", "disabled", "linker_id", "incoming_id", "slave"))
	api.Register("GET", "modbus/device/:id/delete", curd.ParseParamStringId, curd.ApiDelete[internal.Device]())
	api.Register("GET", "modbus/device/:id/enable", curd.ParseParamStringId, curd.ApiDisable[internal.Device](false))
	api.Register("GET", "modbus/device/:id/disable", curd.ParseParamStringId, curd.ApiDisable[internal.Device](true))
	api.Register("GET", "modbus/device/:id/poll", curd.ParseParamStringId, devicePoll)
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
