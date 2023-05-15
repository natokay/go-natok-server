package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/kataras/iris/v12/sessions"
	"natok-server/model/vo"
	"natok-server/service"
)

// ReportController struct 报表 - 控制层
type ReportController struct {
	Ctx     iris.Context
	Session *sessions.Session
	Service *service.ReportService
}

func (c *ReportController) BeforeActivation(b mvc.BeforeActivation) {
	b.Handle("GET", "/dashboard/state", "ClientState")
}

// ClientState 统计客户端状态
func (c *ReportController) ClientState() mvc.Result {
	ret := make(map[string]interface{}, 0)
	ret["stream"] = c.Service.StreamState()
	ret["client"] = c.Service.ClientState()
	ret["port"] = c.Service.PortState()
	ret["protocol"] = c.Service.ProtocolState()
	ret["run"] = c.Service.RunningState()
	return vo.TipResult(ret)
}
