package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/kataras/iris/v12/sessions"
	"natok-server/model"
	"natok-server/model/vo"
	"natok-server/service"
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

// PortList 列表分页
func (c *PortController) PortList() mvc.Result {
	wd := c.Ctx.URLParam("wd")
	page := c.Ctx.URLParamIntDefault("page", 1)
	limit := c.Ctx.URLParamIntDefault("limit", 10)
	ret := c.Service.QueryPort(wd, page, limit)
	return vo.TipResult(ret)
}

// PortGet 详情
func (c *PortController) PortGet() mvc.Result {
	portId := c.Ctx.URLParamInt64Default("portId", 0)
	if ret, err := c.Service.GetPort(portId); err == nil {
		return vo.TipResult(ret)
	} else {
		return vo.TipMsg(err)
	}
}

// PortSave 保存
func (c *PortController) PortSave() mvc.Result {
	item := new(model.NatokPort)
	if err := c.Ctx.ReadJSON(item); err != nil {
		return vo.TipMsg(err)
	}
	if item.AccessKey == "" || item.Intranet == "" || item.PortNum == 0 {
		return vo.TipErrorMsg("parameter error")
	}
	err := c.Service.SavePort(item)
	return vo.TipMsg(err)
}

// PortDelete 删除
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

// PortSwitch 启用或停用
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

// PortValidate 校验
func (c *PortController) PortValidate() mvc.Result {
	portId := c.Ctx.URLParamInt64Default("portId", 0)
	portNum := c.Ctx.URLParamIntDefault("portNum", 0)
	protocol := c.Ctx.URLParamDefault("protocol", "tcp")
	ret := c.Service.ValidatePort(portId, portNum, protocol)
	return vo.TipResult(ret)
}
