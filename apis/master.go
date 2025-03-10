package apis

import (
	"encoding/hex"
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
	api.Register("POST", "modbus/master/:id", curd.ApiUpdate[internal.ModbusMaster]("id", "name", "description", "polling", "polling_interval", "disabled", "linker_id", "incoming_id"))
	api.Register("GET", "modbus/master/:id/delete", curd.ApiDelete[internal.ModbusMaster]())
	api.Register("GET", "modbus/master/:id/enable", curd.ApiDisable[internal.ModbusMaster](false))
	api.Register("GET", "modbus/master/:id/disable", curd.ApiDisable[internal.ModbusMaster](true))
	api.Register("GET", "modbus/master/:id/open", masterOpen)
	api.Register("GET", "modbus/master/:id/close", masterClose)
	api.Register("GET", "modbus/master/:id/read", masterRead)
	api.Register("POST", "modbus/master/:id/write", masterWrite)
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

type masterReadBody struct {
	Slave   uint8  `json:"slave"`
	Code    uint8  `json:"code"`
	Address uint16 `json:"address"`
	Length  uint16 `json:"length"`
}

func masterRead(ctx *gin.Context) {
	c := internal.GetMaster(ctx.Param("id"))
	if c == nil {
		api.Fail(ctx, "master not found")
		return
	}

	var body masterReadBody
	if err := ctx.BindQuery(&body); err != nil {
		api.Error(ctx, err)
		return
	}

	buf, err := c.Read(body.Slave, body.Code, body.Address, body.Length)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, hex.EncodeToString(buf))
}

type masterWriteBody struct {
	Slave   uint8  `json:"slave"`
	Code    uint8  `json:"code"`
	Address uint16 `json:"address"`
	Value   string `json:"value"`
}

func masterWrite(ctx *gin.Context) {
	c := internal.GetMaster(ctx.Param("id"))
	if c == nil {
		api.Fail(ctx, "master not found")
		return
	}

	var body masterWriteBody
	if err := ctx.ShouldBind(&body); err != nil {
		api.Error(ctx, err)
		return
	}

	buf, err := hex.DecodeString(body.Value)

	err = c.Write(body.Slave, body.Code, body.Address, buf)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, nil)
}
