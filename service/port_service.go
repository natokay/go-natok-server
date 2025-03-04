package service

import (
	"errors"
	"natok-server/core"
	"natok-server/dsmapper"
	"natok-server/model"
	"natok-server/util"
	"strconv"
	"time"
)

// PortService struct 端口 - 服务层
type PortService struct {
	Mapper dsmapper.DsMapper
}

// QueryPort PortQuery 列表分页
func (s *PortService) QueryPort(wd string, page, limit int) map[string]interface{} {
	items, total := make([]model.NatokPort, 0), int64(0)
	if wd != "" {
		if client := s.Mapper.ClientQueryByNameOrKey(wd); client != nil {
			items, total = s.Mapper.PortQuery(client.AccessKey, page, limit)
		}
	} else {
		items, total = s.Mapper.PortQuery(wd, page, limit)
	}
	for i, item := range items {
		item.ValidDayCalculate()
		item.WhitelistNilEmpty()
		item.TagNilEmpty()
		items[i] = item
	}
	ret := make(map[string]interface{}, 2)
	ret["items"] = items
	ret["total"] = total
	return ret
}

// ValidatePort 验证端口是否可用
func (s *PortService) ValidatePort(portId int64, portNum int, protocol string) map[string]interface{} {
	ret := make(map[string]interface{}, 2)
	ret["state"] = true
	// 端口号验证
	if portNum > 0 && portNum < 65535 {
		if able := util.PortAvailable(portNum, protocol); able {
			if !s.Mapper.PortExist(portId, portNum, protocol) {
				ret["state"] = false
				return ret
			}
		}
	}
	return ret
}

// GetPort PortGet 详情
func (s *PortService) GetPort(portId int64) (map[string]interface{}, error) {
	ret := make(map[string]interface{}, 2)
	if item := s.Mapper.PortGetById(portId); item != nil {
		item.ValidDayCalculate()
		item.WhitelistNilEmpty()
		item.TagNilEmpty()
		ret["item"] = item
		return ret, nil
	}
	return nil, errors.New("not found")
}

// SavePort 保存端口映射
func (s *PortService) SavePort(item *model.NatokPort) (err error) {
	item.WhitelistNilEmpty()
	item.TagNilEmpty()
	// 过期时间顺延一天
	if item.ValidDay <= 0 {
		item.ExpireAt = util.NowDayStart().Add(8 * 24 * time.Hour)
	} else {
		item.ExpireAt = util.NowDayStart().Add(time.Duration(item.ValidDay+1) * 24 * time.Hour)
	}
	// 事务保证一致性
	return s.Mapper.Transaction(func() error {
		if item.PortId <= 0 {
			//直接插入
			item.CreateAt = time.Now()
			item.PortSign = util.Md5(util.PreTs(util.UUID() + item.AccessKey + strconv.Itoa(item.PortNum)))
			item.State = 1
			item.Apply = 0
			item.Enabled = 1
			if err = s.Mapper.PortSaveUp(item); err != nil {
				return err
			}
			if err = s.SwitchPort(item.PortId, item.AccessKey, 1); err != nil {
				return err
			}
		} else {
			port := s.Mapper.PortGetById(item.PortId)
			if nil == port {
				return errors.New("not found")
			}
			if port.Enabled == 1 {
				return errors.New("cannot be modified")
			}
			port.PortScope = item.PortScope
			port.PortNum = item.PortNum
			port.Intranet = item.Intranet
			port.Protocol = item.Protocol
			port.ExpireAt = item.ExpireAt
			port.Remark = item.Remark
			port.Whitelist = item.Whitelist
			port.Tag = item.Tag
			port.Modified = time.Now()
			if err = s.Mapper.PortSaveUp(port); err != nil {
				return err
			}
		}
		// 更新开放名单
		if err = PortWhitelist(item, s.Mapper); err != nil {
			return err
		}
		return nil
	})
}

// SwitchPort 启用或停用
func (s *PortService) SwitchPort(portId int64, accessKey string, enabled int8) (err error) {
	port := s.Mapper.PortGetById(portId)
	if nil != port && port.AccessKey == accessKey {
		if port.State <= 0 {
			return errors.New("client not enabled")
		}
		port.Enabled = enabled
		port.Modified = time.Now()
		port.WhitelistNilEmpty()
		port.TagNilEmpty()
		// 事务保证一致性
		return s.Mapper.Transaction(func() error {
			if err = s.Mapper.PortSaveUp(port); err != nil {
				return err
			}
			// 端口映射：绑定与解绑
			if err = SwitchPortMapping(&core.PortMapping{
				AccessKey: port.AccessKey, PortSign: port.PortSign,
				PortScope: port.PortScope,
				PortNum:   port.PortNum, Intranet: port.Intranet,
				Protocol:  port.Protocol,
				Whitelist: port.Whitelist,
			}, int(2-enabled)); err != nil {
				return err
			}
			// 更新开放名单
			if err = PortWhitelist(port, s.Mapper); err != nil {
				return err
			}
			return nil
		})
	}
	return errors.New("not found")
}

// DeletePort 删除
func (s *PortService) DeletePort(portId int64, accessKey string) (err error) {
	port := s.Mapper.PortGetById(portId)
	if nil == port || port.AccessKey != accessKey {
		return errors.New("not found")
	}
	port.Deleted = 1
	port.Modified = time.Now()
	port.WhitelistNilEmpty()
	port.TagNilEmpty()
	// 事务保证一致性
	return s.Mapper.Transaction(func() error {
		if err = s.Mapper.PortSaveUp(port); err != nil {
			return err
		}
		// 端口映射：解绑
		if err = SwitchPortMapping(&core.PortMapping{
			AccessKey: port.AccessKey, PortSign: port.PortSign,
			PortScope: port.PortScope,
			PortNum:   port.PortNum, Intranet: port.Intranet,
			Protocol:  port.Protocol,
			Whitelist: port.Whitelist,
		}, 2); err != nil {
			return err
		}
		return nil
	})
}

// SwitchPortMapping 启停端口映射
// opt=1 启用 opt=2 停用
func SwitchPortMapping(mapping *core.PortMapping, opt int) error {
	var err error
	switch opt {
	case 1:
		err = core.BindPort(mapping)
	case 2:
		err = core.UnBindPort(mapping)
	}
	return err
}
