package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/kataras/iris/v12/sessions"
	"natok-server/model/vo"
	"natok-server/service"
)

// ClientController struct 客户端 - 控制层
type ClientController struct {
	Ctx     iris.Context
	Session *sessions.Session
	Service *service.ClientService
}

func (c *ClientController) BeforeActivation(b mvc.BeforeActivation) {
	b.Handle("GET", "/client/list", "ClientList")
	b.Handle("GET", "/client/get", "ClientGet")
	b.Handle("PUT", "/client/save", "ClientSave")
	b.Handle("POST", "/client/switch", "ClientSwitch")
	b.Handle("POST", "/client/validate", "ClientValidate")
	b.Handle("DELETE", "/client/del", "ClientDelete")
	b.Handle("GET", "/client/keys", "ClientKey")
}

// ClientList 列表分页
func (c *ClientController) ClientList() mvc.Result {
	wd := c.Ctx.URLParam("wd")
	page := c.Ctx.URLParamIntDefault("page", 1)
	limit := c.Ctx.URLParamIntDefault("limit", 10)
	ret := c.Service.ClientQuery(wd, page, limit)
	return vo.TipResult(ret)
}

// ClientGet 详情
func (c *ClientController) ClientGet() mvc.Result {
	clientId := c.Ctx.URLParamInt64Default("clientId", 0)
	//非正常情况，返回错误消息
	if clientId <= 0 {
		return vo.TipErrorMsg("parameter error")
	}
	ret := c.Service.ClientGet(clientId)
	return vo.TipResult(ret)
}

// ClientSave 保存
func (c *ClientController) ClientSave() mvc.Result {
	clientId := c.Ctx.URLParamInt64Default("clientId", 0)
	accessKey := c.Ctx.URLParam("accessKey")
	clientName := c.Ctx.URLParam("clientName")
	err := c.Service.ClientSave(clientId, clientName, accessKey)
	return vo.TipMsg(err)
}

// ClientDelete 删除
func (c *ClientController) ClientDelete() mvc.Result {
	clientId := c.Ctx.URLParamInt64Default("clientId", 0)
	accessKey := c.Ctx.URLParam("accessKey")
	//非正常情况，返回错误消息
	if clientId <= 0 || accessKey == "" {
		return vo.TipErrorMsg("parameter error")
	}
	err := c.Service.ClientDelete(clientId, accessKey)
	return vo.TipMsg(err)
}

// ClientKey 获取客户端的name+accessKey的集合
func (c *ClientController) ClientKey() mvc.Result {
	ret := c.Service.ClientKeys()
	return vo.TipResult(ret)
}

// ClientValidate 校验
func (c *ClientController) ClientValidate() mvc.Result {
	clientId := c.Ctx.URLParamInt64Default("clientId", 0)
	value := c.Ctx.URLParam("value")
	level := c.Ctx.URLParamInt32Default("type", 0)
	ret := c.Service.ClientValidate(clientId, value, level)
	return vo.TipResult(ret)
}

// ClientSwitch 启用与停用
func (c *ClientController) ClientSwitch() mvc.Result {
	clientId := c.Ctx.URLParamInt64Default("clientId", 0)
	enabled := c.Ctx.URLParamIntDefault("enabled", 0)
	accessKey := c.Ctx.URLParam("accessKey")
	//非正常情况，返回错误消息
	if clientId <= 0 || accessKey == "" || 0 > enabled || enabled > 1 {
		return vo.TipErrorMsg("parameter error")
	}
	err := c.Service.ClientSwitch(clientId, accessKey, int8(enabled))
	return vo.TipMsg(err)
}
