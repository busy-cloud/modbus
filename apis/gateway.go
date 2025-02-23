package apis

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
	"github.com/busy-cloud/modbus/internal"
	"github.com/gin-gonic/gin"
)

func init() {
	api.Register("POST", "modbus/gateway/count", curd.ApiCount[internal.Gateway]())
	api.Register("POST", "modbus/gateway/search", curd.ApiSearch[internal.Gateway]())
	api.Register("GET", "modbus/gateway/list", curd.ApiList[internal.Gateway]())
	api.Register("POST", "modbus/gateway/create", curd.ApiCreate[internal.Gateway]())
	api.Register("GET", "modbus/gateway/:id", curd.ParseParamStringId, curd.ApiGet[internal.Gateway]())
	api.Register("POST", "modbus/gateway/:id", curd.ParseParamStringId, curd.ApiUpdate[internal.Gateway]("id", "name", "disabled"))
	api.Register("GET", "modbus/gateway/:id/delete", curd.ParseParamStringId, curd.ApiDelete[internal.Gateway]())
	api.Register("GET", "modbus/gateway/:id/enable", curd.ParseParamStringId, curd.ApiDisable[internal.Gateway](false))
	api.Register("GET", "modbus/gateway/:id/disable", curd.ParseParamStringId, curd.ApiDisable[internal.Gateway](true))
	api.Register("GET", "modbus/gateway/:id/open", curd.ParseParamStringId, gatewayOpen)
	api.Register("GET", "modbus/gateway/:id/close", curd.ParseParamStringId, gatewayClose)
}

func gatewayOpen(ctx *gin.Context) {
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

func gatewayClose(ctx *gin.Context) {
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
