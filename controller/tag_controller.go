package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/kataras/iris/v12/sessions"
	"natok-server/model"
	"natok-server/model/vo"
	"natok-server/service"
)

// TagController struct  - 控制层
type TagController struct {
	Ctx     iris.Context
	Session *sessions.Session
	Service *service.TagService
}

func (c *TagController) BeforeActivation(b mvc.BeforeActivation) {
	b.Handle("GET", "/tag/list", "TagList")
	b.Handle("GET", "/tag/get", "TagGet")
	b.Handle("PUT", "/tag/save", "TagSave")
	b.Handle("POST", "/tag/switch", "TagSwitch")
	b.Handle("DELETE", "/tag/del", "TagDelete")
}

// TagList 列表分页
func (c *TagController) TagList() mvc.Result {
	wd := c.Ctx.URLParam("wd")
	page := c.Ctx.URLParamIntDefault("page", 1)
	limit := c.Ctx.URLParamIntDefault("limit", 10)
	ret := c.Service.QueryPageTag(wd, page, limit)
	return vo.TipResult(ret)
}

// TagGet 详情
func (c *TagController) TagGet() mvc.Result {
	tagId := c.Ctx.URLParamInt64Default("tagId", 0)
	if ret, err := c.Service.GetTag(tagId); err == nil {
		return vo.TipResult(ret)
	} else {
		return vo.TipMsg(err)
	}
}

// TagSave 保存
func (c *TagController) TagSave() mvc.Result {
	item := new(model.NatokTag)
	if err := c.Ctx.ReadJSON(item); err != nil {
		return vo.TipMsg(err)
	}
	if item.TagName == "" {
		return vo.TipErrorMsg("parameter error")
	}
	err := c.Service.SaveTag(item)
	return vo.TipMsg(err)
}

// TagSwitch 启用或停用
func (c *TagController) TagSwitch() mvc.Result {
	tagId := c.Ctx.URLParamInt64Default("tagId", 0)
	enabled := c.Ctx.URLParamIntDefault("enabled", 0)
	//非正常情况，返回错误消息
	if tagId <= 0 {
		return vo.TipErrorMsg("parameter error")
	}
	err := c.Service.SwitchTag(tagId, int8(enabled))
	return vo.TipMsg(err)
}

// TagDelete 删除
func (c *TagController) TagDelete() mvc.Result {
	tagId := c.Ctx.URLParamInt64Default("tagId", 0)
	//非正常情况，返回错误消息
	if tagId <= 0 {
		return vo.TipErrorMsg("parameter error")
	}
	err := c.Service.DeleteTag(tagId)
	return vo.TipMsg(err)
}
