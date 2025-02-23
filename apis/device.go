package apis

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/modbus/internal"
	"github.com/gin-gonic/gin"
)

func init() {
	api.Register("POST", "modbus/device/count", curd.ApiCount[internal.Device]())
	api.Register("POST", "modbus/device/search", curd.ApiSearch[internal.Device]())
	api.Register("GET", "modbus/device/list", curd.ApiList[internal.Device]())
	api.Register("POST", "modbus/device/create", curd.ApiCreate[internal.Device]())
	api.Register("GET", "modbus/device/:id", curd.ParseParamStringId, curd.ApiGet[internal.Device]())
	api.Register("POST", "modbus/device/:id", curd.ParseParamStringId, curd.ApiUpdate[internal.Device]("id", "name", "line", "gateway_id", "product_id", "disabled"))
	api.Register("GET", "modbus/device/:id/delete", curd.ParseParamStringId, curd.ApiDelete[internal.Device]())
	api.Register("GET", "modbus/device/:id/enable", curd.ParseParamStringId, curd.ApiDisable[internal.Device](false))
	api.Register("GET", "modbus/device/:id/disable", curd.ParseParamStringId, curd.ApiDisable[internal.Device](true))
	api.Register("GET", "modbus/device/:id/status", curd.ParseParamStringId, deviceStatus)
	api.Register("GET", "modbus/device/:id/realtime_status", curd.ParseParamStringId, deviceRealtimeStatus)
	api.Register("GET", "modbus/device/:id/setting", curd.ParseParamStringId, deviceSettingGet)
	api.Register("POST", "modbus/device/:id/setting", curd.ParseParamStringId, deviceSettingSet)
}

func deviceRealtimeStatus(ctx *gin.Context) {
	var device internal.Device
	has, err := db.Engine.ID(ctx.Param("id")).Get(&device)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	if !has {
		api.Fail(ctx, "device not found")
		return
	}

	gateway := internal.GetGateway(device.GatewayId)
	if gateway == nil {
		api.Fail(ctx, "gateway not online")
		return
	}

	d := gateway.GetDevice(device.Id)
	if d == nil {
		api.Fail(ctx, "device not found")
		return
	}

	st, err := d.Status(true)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, st)
}

func deviceStatus(ctx *gin.Context) {
	var device internal.Device
	has, err := db.Engine.ID(ctx.Param("id")).Get(&device)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	if !has {
		api.Fail(ctx, "device not found")
		return
	}

	gateway := internal.GetGateway(device.GatewayId)
	if gateway == nil {
		api.Fail(ctx, "gateway not online")
		return
	}

	d := gateway.GetDevice(device.Id)
	if d == nil {
		api.Fail(ctx, "device not found")
		return
	}

	api.OK(ctx, d.GetStatus())
}

type deviceSettingGetQuery struct {
	Offset uint8
	Length uint8
}

type deviceSettingSetBody struct {
	Offset uint8
	Value  uint16
}

func deviceSettingGet(ctx *gin.Context) {
	var device internal.Device
	has, err := db.Engine.ID(ctx.Param("id")).Get(&device)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	if !has {
		api.Fail(ctx, "device not found")
		return
	}

	gateway := internal.GetGateway(device.GatewayId)
	if gateway == nil {
		api.Fail(ctx, "gateway not online")
		return
	}

	d := gateway.GetDevice(device.Id)
	if d == nil {
		api.Fail(ctx, "device not found")
		return
	}

	var query deviceSettingGetQuery
	err = ctx.BindQuery(&query)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	st, err := d.GetSetting(query.Offset, query.Length)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, st)
}

func deviceSettingSet(ctx *gin.Context) {
	var device internal.Device
	has, err := db.Engine.ID(ctx.Param("id")).Get(&device)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	if !has {
		api.Fail(ctx, "device not found")
		return
	}

	gateway := internal.GetGateway(device.GatewayId)
	if gateway == nil {
		api.Fail(ctx, "gateway not online")
		return
	}

	d := gateway.GetDevice(device.Id)
	if d == nil {
		api.Fail(ctx, "device not found")
		return
	}

	var query deviceSettingSetBody
	err = ctx.BindJSON(&query)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	err = d.SetSetting(query.Offset, query.Value)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, 0)
}
