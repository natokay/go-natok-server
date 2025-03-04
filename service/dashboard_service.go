package service

import (
	"natok-server/core"
	"natok-server/dsmapper"
	"sync/atomic"
)

// ReportService struct 报表 - 服务层
type ReportService struct {
	Mapper dsmapper.DsMapper
}

// StreamState 统计流入流出量
func (s *ReportService) StreamState() map[string]interface{} {
	ret := make(map[string]interface{}, 0)
	ret["input"] = map[string]interface{}{"name": "流入", "data": []int{100, 120, 1, 134, 105, 160, 165}}
	ret["output"] = map[string]interface{}{"name": "流出", "data": []int{120, 82, 91, 154, 162, 140, 145}}
	ret["date"] = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	return ret
}

// ClientState 统计客户端状态
func (s *ReportService) ClientState() map[string]interface{} {
	ret := make(map[string]interface{}, 0)
	state := s.Mapper.ClientGroupByState()
	if state["在线"] == nil {
		state["在线"] = 0
	}
	if state["离线"] == nil {
		state["离线"] = 0
	}
	for k, v := range state {
		ret[k] = v
	}
	enabled := s.Mapper.ClientGroupByEnabled()
	if enabled["启用"] == nil {
		enabled["启用"] = 0
	}
	if enabled["停用"] == nil {
		enabled["停用"] = 0
	}
	for k, v := range enabled {
		ret[k] = v
	}
	return ret
}

// PortState 统计端口状态
func (s *ReportService) PortState() map[string]interface{} {
	ret := make(map[string]interface{}, 0)
	actCount := int64(0)
	core.ClientManage.Range(func(_, cm any) bool {
		client := cm.(*core.ClientBlocking)
		count := core.GetLen(&client.PortListener)
		atomic.AddInt64(&actCount, int64(count))
		return true
	})
	ret["活跃"] = actCount
	ret["未活动"] = s.Mapper.PortCountTotal("") - actCount
	ret["未过期"] = s.Mapper.PortExpiredCount(false)
	ret["已过期"] = s.Mapper.PortExpiredCount(true)
	return ret
}

// ProtocolState 统计端口协议状态
func (s *ReportService) ProtocolState() map[string]interface{} {
	ret := s.Mapper.PortGroupByProtocol()
	if nil == ret["TCP"] {
		ret["TCP"] = 0
	}
	if nil == ret["SSH"] {
		ret["SSH"] = 0
	}
	if nil == ret["HTTP"] {
		ret["HTTP"] = 0
	}
	if nil == ret["HTTPS"] {
		ret["HTTPS"] = 0
	}
	if nil == ret["DataBase"] {
		ret["DataBase"] = 0
	}
	if nil == ret["Telnet"] {
		ret["Telnet"] = 0
	}
	if nil == ret["Desktop"] {
		ret["Desktop"] = 0
	}
	return ret
}

// RunningState 统计运行状态
func (s *ReportService) RunningState() []interface{} {
	ret := make([]any, 0)
	core.ClientManage.Range(func(key, cm any) bool {
		client := cm.(*core.ClientBlocking)
		if cn, ifCN := core.ConnectManage.Load(key); cn != nil && ifCN {
			blocking := cn.(*core.ConnectBlocking)
			client.PortListener.Range(func(sign, pm any) bool {
				mapping := pm.(*core.PortMapping)
				item := make(map[string]interface{}, 0)
				if signs, ifSign := blocking.PortSignMap.Load(sign); signs != nil && ifSign {
					item["cn"] = len(signs.([]string))
				} else {
					item["cn"] = 0
				}
				item["port"] = mapping.PortNum
				item["protocol"] = mapping.Protocol
				ret = append(ret, item)
				return true
			})
		}
		return true
	})
	return ret
}
