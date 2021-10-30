package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/kataras/iris/v12/sessions"
	"go-natok-server/model/vo"
	"go-natok-server/service"
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

// ClientList 获取C端列表
func (c *ClientController) ClientList() mvc.Result {
	wd := c.Ctx.URLParam("wd")
	sort := c.Ctx.URLParam("sort")
	page := c.Ctx.URLParamIntDefault("page", 1)
	limit := c.Ctx.URLParamIntDefault("limit", 10)
	ret := c.Service.QueryClient(wd, sort, page, limit)
	return vo.TipResult(ret)
}

// ClientGet 获取单个C端信息
func (c *ClientController) ClientGet() mvc.Result {
	clientId := c.Ctx.URLParamInt64Default("clientId", 0)
	//非正常情况，返回错误消息
	if clientId <= 0 {
		return vo.TipErrorMsg("parameter error")
	}
	ret := c.Service.GetClient(clientId)
	return vo.TipResult(ret)
}

// ClientSave 保存单个C端信息
func (c *ClientController) ClientSave() mvc.Result {
	clientId := c.Ctx.URLParamInt64Default("clientId", 0)
	accessKey := c.Ctx.URLParam("accessKey")
	clientName := c.Ctx.URLParam("clientName")
	err := c.Service.SaveClient(clientId, clientName, accessKey)
	return vo.TipMsg(err)
}

// ClientKey 获取C端的name+accessKey的集合
func (c *ClientController) ClientKey() mvc.Result {
	ret := c.Service.ClientKeys()
	return vo.TipResult(ret)
}
func (c *ClientController) ClientValidate() mvc.Result {
	clientId := c.Ctx.URLParamInt64Default("clientId", 0)
	value := c.Ctx.URLParam("value")
	level := c.Ctx.URLParamInt32Default("type", 0)
	ret := c.Service.ValidateClient(clientId, value, level)
	return vo.TipResult(ret)
}

// ClientSwitch 启用与停用切换
func (c *ClientController) ClientSwitch() mvc.Result {
	clientId := c.Ctx.URLParamInt64Default("clientId", 0)
	enabled := c.Ctx.URLParamIntDefault("enabled", 0)
	accessKey := c.Ctx.URLParam("accessKey")
	//非正常情况，返回错误消息
	if clientId <= 0 || accessKey == "" || 0 > enabled || enabled > 1 {
		return vo.TipErrorMsg("parameter error")
	}
	err := c.Service.SwitchClient(clientId, accessKey, int8(enabled))
	return vo.TipMsg(err)
}

// ClientDelete 删除端口映射
func (c *ClientController) ClientDelete() mvc.Result {
	clientId := c.Ctx.URLParamInt64Default("clientId", 0)
	accessKey := c.Ctx.URLParam("accessKey")
	//非正常情况，返回错误消息
	if clientId <= 0 || accessKey == "" {
		return vo.TipErrorMsg("parameter error")
	}
	err := c.Service.DeleteClient(clientId, accessKey)
	return vo.TipMsg(err)
}
