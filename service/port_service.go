package service

import (
	"errors"
	"go-natok-server/core"
	"go-natok-server/dsmapper"
	"go-natok-server/model"
	"go-natok-server/support"
	"go-natok-server/util"
	"strconv"
	"time"
)

// PortService struct 端口 - 服务层
type PortService struct {
	Mapper dsmapper.DsMapper
}

// QueryPort PortQuery 查询并封装端口映射分页数据
func (s *PortService) QueryPort(wd, sort string, page, limit int) map[string]interface{} {
	var items, total = make([]model.NatokPort, 0), int64(0)
	if wd != "" {
		if client := s.Mapper.ClientQueryByNameOrKey(wd); client != nil {
			items, total = s.Mapper.PortQuery(client.AccessKey, sort, page, limit)
		}
	} else {
		items, total = s.Mapper.PortQuery(wd, sort, page, limit)
	}
	for idx, item := range items {
		items[idx].ValidDay = item.GetValidDay()
	}
	var ret = make(map[string]interface{}, 2)
	ret["items"] = items
	ret["total"] = total
	return ret
}

// ValidatePort 端口映射，存在验证
func (s *PortService) ValidatePort(portId int64, value string, level int32) map[string]interface{} {
	var ret = make(map[string]interface{}, 2)
	ret["state"] = true
	if level == 1 { //端口号验证
		if portNum, err := strconv.Atoi(value); err == nil && portNum > 0 && portNum < 65535 {
			pid, err := support.PortCheckup(portNum)
			if err != nil || pid != -1 {
				port := s.Mapper.PortGetById(portId)
				if port != nil && port.PortNum != portNum {
					return ret
				}
			}
			ret["state"] = pid != -1 || s.Mapper.PortExist(portId, value)
		}
	}
	if level == 2 { //域名验证
		ret["state"] = s.Mapper.PortExist(portId, value)
	}
	return ret
}

// GetPort PortGet 获取端口映射
func (s *PortService) GetPort(portId int64) map[string]interface{} {
	var ret = make(map[string]interface{}, 2)
	port := s.Mapper.PortGetById(portId)
	port.ValidDay = port.GetValidDay()
	ret["item"] = port
	return ret
}

// SavePort 保存端口映射
func (s *PortService) SavePort(item *model.NatokPort) error {
	if item != nil && item.AccessKey != "" {
		//过期时间顺延一天
		if item.ValidDay <= 0 {
			item.ExpireAt = util.NowDayStart().Add(8 * 24 * time.Hour)
		} else {
			item.ExpireAt = util.NowDayStart().Add(time.Duration(item.ValidDay+1) * 24 * time.Hour)
		}
		//添加端口映射
		if item.PortId <= 0 && item.AccessKey != "" {
			item.CreateAt = time.Now()
			item.Sign = util.Md5(time.Now().String() + util.UUID() + item.AccessKey + strconv.Itoa(item.PortNum))
			item.State = 1
			item.Apply = 1
			item.Enabled = 1
		}
		//更新端口映射
		if item.PortId > 0 {
			port := s.Mapper.PortGetById(item.PortId)
			port.Domain = item.Domain
			port.PortNum = item.PortNum
			port.Intranet = item.Intranet
			port.Protocol = item.Protocol
			port.ExpireAt = item.ExpireAt
			port.Remark = item.Remark
			port.Modified = time.Now()
			item = port
		}
		//TODO 事务保证一致性
		err := s.Mapper.Transaction()
		if nil == err {
			err = s.Mapper.PortSaveUp(item)
		}
		// 端口映射：绑定与更新绑定
		if nil == err {
			err = SwitchPortMapping(core.PortMapping{
				AccessKey: item.AccessKey, Sign: item.Sign,
				Port: item.PortNum, Intranet: item.Intranet,
				Domain: item.Domain, Protocol: item.Protocol,
			}, 3)
		}
		if nil == err {
			s.Mapper.Commit()
		} else {
			s.Mapper.Rollback()
		}
		return err
	}
	return nil
}

// SwitchPort 启用或停用端口映射
func (s *PortService) SwitchPort(portId int64, accessKey string, enabled int8) error {
	port := s.Mapper.PortGetById(portId)
	if nil != port && port.AccessKey == accessKey {
		if port.State <= 0 {
			return errors.New("client not enabled")
		}
		port.Enabled = enabled
		port.Modified = time.Now()

		//TODO 事务保证一致性
		err := s.Mapper.Transaction()
		if nil == err {
			err = s.Mapper.PortSaveUp(port)
		}
		// 端口映射：绑定与解绑
		if nil == err {
			err = SwitchPortMapping(core.PortMapping{
				AccessKey: port.AccessKey, Sign: port.Sign,
				Port: port.PortNum, Intranet: port.Intranet,
				Domain: port.Domain, Protocol: port.Protocol,
			}, int(2-enabled))
		}
		if nil == err {
			s.Mapper.Commit()
		} else {
			s.Mapper.Rollback()
		}
		return err
	}
	return errors.New("not found")
}

// DeletePort 标记删除端口映射
func (s *PortService) DeletePort(portId int64, accessKey string) error {
	port := s.Mapper.PortGetById(portId)
	if nil != port && port.AccessKey == accessKey {
		port.Deleted = 1
		port.Modified = time.Now()

		//TODO 事务保证一致性
		err := s.Mapper.Transaction()
		if nil == err {
			err = s.Mapper.PortSaveUp(port)
		}
		// 端口映射：解绑
		err = SwitchPortMapping(core.PortMapping{
			AccessKey: port.AccessKey, Sign: port.Sign,
			Port: port.PortNum, Intranet: port.Intranet,
			Domain: port.Domain, Protocol: port.Protocol,
		}, 2)
		if nil == err {
			s.Mapper.Commit()
		} else {
			s.Mapper.Rollback()
		}
		return err
	}
	return errors.New("not found")
}

// SwitchPortMapping 启停端口映射
// opt=1 启用 opt=2 停用 opt=3 重启
func SwitchPortMapping(mapping core.PortMapping, opt int) error {
	var err error
	switch opt {
	case 1:
		err = core.BindPort(mapping)
	case 2:
		err = core.UnBindPort(mapping)
	case 3:
		err = core.UnBindPort(mapping)
		if nil == err {
			err = core.BindPort(mapping)
		}
	}
	return err
}
