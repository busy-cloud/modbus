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
	api.Register("GET", "modbus/master/:id", curd.ApiGet[internal.ModbusMaster]())
	api.Register("POST", "modbus/master/:id", curd.ApiUpdate[internal.ModbusMaster]("id", "name", "description", "polling", "polling_interval", "disabled", "linker_id", "incoming_id", "slave"))
	api.Register("GET", "modbus/master/:id/delete", curd.ApiDelete[internal.ModbusMaster]())
	api.Register("GET", "modbus/master/:id/enable", curd.ApiDisable[internal.ModbusMaster](false))
	api.Register("GET", "modbus/master/:id/disable", curd.ApiDisable[internal.ModbusMaster](true))
	api.Register("GET", "modbus/master/:id/open", masterOpen)
	api.Register("GET", "modbus/master/:id/close", masterClose)
}

func masterOpen(ctx *gin.Context) {
	c := internal.GetMaster(ctx.Param("id"))
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
	c := internal.GetMaster(ctx.Param("id"))
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
