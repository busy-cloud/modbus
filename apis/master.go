package apis

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
	"github.com/busy-cloud/modbus/internal"
	"github.com/gin-gonic/gin"
)

func init() {
	api.Register("GET", "modbus/master/:linker/:incoming/open", curd.ParseParamStringId, masterOpen)
	api.Register("GET", "modbus/master/:linker/open", curd.ParseParamStringId, masterOpen)
	api.Register("GET", "modbus/master/:linker/:incoming/close", curd.ParseParamStringId, masterClose)
	api.Register("GET", "modbus/master/:linker/close", curd.ParseParamStringId, masterClose)
}

func masterOpen(ctx *gin.Context) {
	c := internal.GetMaster(ctx.Param("linker"), ctx.Param("incoming"))
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
	c := internal.GetMaster(ctx.Param("linker"), ctx.Param("incoming"))
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
