package apis

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
	"github.com/busy-cloud/modbus/internal"
	"github.com/gin-gonic/gin"
)

func init() {
	api.Register("POST", "modbus/master/count", curd.ApiCount[internal.Master]())
	api.Register("POST", "modbus/master/search", curd.ApiSearch[internal.Master]())
	api.Register("GET", "modbus/master/list", curd.ApiList[internal.Master]())
	api.Register("POST", "modbus/master/create", curd.ApiCreate[internal.Master]())
	api.Register("GET", "modbus/master/:id", curd.ApiGet[internal.Master]())
	api.Register("POST", "modbus/master/:id", curd.ApiUpdate[internal.Master]("id", "name", "description", "product_id", "disabled", "linker_id", "incoming_id", "slave"))
	api.Register("GET", "modbus/master/:id/delete", curd.ApiDelete[internal.Master]())
	api.Register("GET", "modbus/master/:id/enable", curd.ApiDisable[internal.Master](false))
	api.Register("GET", "modbus/master/:id/disable", curd.ApiDisable[internal.Master](true))
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
