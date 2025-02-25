package apis

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
	"github.com/busy-cloud/modbus/internal"
	"github.com/gin-gonic/gin"
)

func init() {
	api.Register("POST", "modbus/master/count", curd.ApiCount[internal.ModbusMaster]())
	api.Register("POST", "modbus/master/search", curd.ApiSearch[internal.ModbusMaster]())
	api.Register("GET", "modbus/master/list", curd.ApiList[internal.ModbusMaster]())
	api.Register("POST", "modbus/master/create", curd.ApiCreate[internal.ModbusMaster]())
	api.Register("GET", "modbus/master/:id", curd.ParseParamStringId, curd.ApiGet[internal.ModbusMaster]())
	api.Register("POST", "modbus/master/:id", curd.ParseParamStringId, curd.ApiUpdate[internal.ModbusMaster]("id", "name", "disabled"))
	api.Register("GET", "modbus/master/:id/delete", curd.ParseParamStringId, curd.ApiDelete[internal.ModbusMaster]())
	api.Register("GET", "modbus/master/:id/enable", curd.ParseParamStringId, curd.ApiDisable[internal.ModbusMaster](false))
	api.Register("GET", "modbus/master/:id/disable", curd.ParseParamStringId, curd.ApiDisable[internal.ModbusMaster](true))
	api.Register("GET", "modbus/master/:id/open", curd.ParseParamStringId, masterOpen)
	api.Register("GET", "modbus/master/:id/close", curd.ParseParamStringId, masterClose)
}

func masterOpen(ctx *gin.Context) {
	c := internal.GetGateway(ctx.Param("id"))
	if c == nil {
		api.Fail(ctx, "找不到连接")
		return
	}

	err := c.Open()
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, nil)
}

func masterClose(ctx *gin.Context) {
	c := internal.GetGateway(ctx.Param("id"))
	if c == nil {
		api.Fail(ctx, "找不到连接")
		return
	}

	err := c.Close()
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, nil)
}
