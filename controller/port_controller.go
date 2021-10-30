package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/kataras/iris/v12/sessions"
	"go-natok-server/model"
	"go-natok-server/model/vo"
	"go-natok-server/service"
)

// PortController struct 端口 - 控制层
type PortController struct {
	Ctx     iris.Context
	Session *sessions.Session
	Service *service.PortService
}

func (c *PortController) BeforeActivation(b mvc.BeforeActivation) {
	b.Handle("GET", "/port/list", "PortList")
	b.Handle("GET", "/port/get", "PortGet")
	b.Handle("PUT", "/port/save", "PortSave")
	b.Handle("POST", "/port/validate", "PortValidate")
	b.Handle("POST", "/port/switch", "PortSwitch")
	b.Handle("DELETE", "/port/del", "PortDelete")
}

// PortList 获取端口列表
func (c *PortController) PortList() mvc.Result {
	wd := c.Ctx.URLParam("wd")
	sort := c.Ctx.URLParam("sort")
	page := c.Ctx.URLParamIntDefault("page", 1)
	limit := c.Ctx.URLParamIntDefault("limit", 10)
	ret := c.Service.QueryPort(wd, sort, page, limit)
	return vo.TipResult(ret)
}

// PortValidate 端口校验
func (c *PortController) PortValidate() mvc.Result {
	portId := c.Ctx.URLParamInt64Default("portId", 0)
	value := c.Ctx.URLParam("value")
	level := c.Ctx.URLParamInt32Default("type", 0)
	ret := c.Service.ValidatePort(portId, value, level)
	return vo.TipResult(ret)
}

// PortGet 获取端口信息
func (c *PortController) PortGet() mvc.Result {
	portId := c.Ctx.URLParamInt64Default("portId", 0)
	ret := c.Service.GetPort(portId)
	return vo.TipResult(ret)
}

// PortSave 保存单个C端信息
func (c *PortController) PortSave() mvc.Result {
	item := new(model.NatokPort)
	err := c.Ctx.ReadForm(item)
	if err == nil {
		err = c.Service.SavePort(item)
	}
	return vo.TipMsg(err)
}

// PortSwitch 启用与停用切换
func (c *PortController) PortSwitch() mvc.Result {
	portId := c.Ctx.URLParamInt64Default("portId", 0)
	enabled := c.Ctx.URLParamIntDefault("enabled", 0)
	accessKey := c.Ctx.URLParam("accessKey")
	//非正常情况，返回错误消息
	if portId <= 0 || accessKey == "" {
		return vo.TipErrorMsg("parameter error")
	}
	err := c.Service.SwitchPort(portId, accessKey, int8(enabled))
	return vo.TipMsg(err)
}

// PortDelete 删除端口映射
func (c *PortController) PortDelete() mvc.Result {
	portId := c.Ctx.URLParamInt64Default("portId", 0)
	accessKey := c.Ctx.URLParam("accessKey")
	//非正常情况，返回错误消息
	if portId <= 0 || accessKey == "" {
		return vo.TipErrorMsg("parameter error")
	}
	err := c.Service.DeletePort(portId, accessKey)
	return vo.TipMsg(err)
}
