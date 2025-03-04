package service

import (
	"errors"
	"natok-server/core"
	"natok-server/dsmapper"
	"natok-server/model"
	"natok-server/util"
	"time"
)

// TagService struct 标签 - 服务层
type TagService struct {
	Mapper dsmapper.DsMapper
}

// QueryPageTag 列表分页
func (s *TagService) QueryPageTag(wd string, page, limit int) map[string]interface{} {
	items, total := make([]model.NatokTag, 0), int64(0)
	items, total = s.Mapper.TagQuery(wd, page, limit)
	for i, item := range items {
		item.WhitelistNilEmpty()
		items[i] = item
	}
	ret := make(map[string]interface{}, 2)
	ret["items"] = items
	ret["total"] = total
	return ret
}

// GetTag 获取
func (s *TagService) GetTag(tagId int64) (map[string]interface{}, error) {
	ret := make(map[string]interface{}, 2)
	if item := s.Mapper.TagGetById(tagId); item != nil {
		item.WhitelistNilEmpty()
		ret["item"] = item
		return ret, nil
	}
	return nil, errors.New("not found")
}

// SaveTag 保存
func (s *TagService) SaveTag(item *model.NatokTag) (err error) {
	if item != nil {
		// 检查名称是否已存在
		if byName := s.Mapper.TagGetByName(item.TagName); byName != nil {
			if item.TagId <= 0 || item.TagId != byName.TagId {
				return errors.New("tag name already exists")
			}
		}
		// 添加标签映射
		if item.TagId <= 0 {
			item.Created = time.Now()
			item.Enabled = 1
			item.Deleted = 0
		} else {
			item.Modified = time.Now()
		}
		// 事务保证一致性
		return s.Mapper.Transaction(func() error {
			if err = s.Mapper.TagSaveUp(item); err != nil {
				return err
			}
			if err = s.refresh(item.TagId); err != nil {
				return err
			}
			return nil
		})
	}
	return err
}

// DeleteTag 标记删除标签映射
func (s *TagService) DeleteTag(tagId int64) (err error) {
	tag := s.Mapper.TagGetById(tagId)
	if nil != tag {
		tag.Deleted = 1
		tag.Enabled = 0
		tag.Modified = time.Now()
		// 事务保证一致性
		return s.Mapper.Transaction(func() error {
			if err = s.Mapper.TagSaveUp(tag); err != nil {
				return err
			}
			if err = s.refresh(tagId); err != nil {
				return err
			}
			return nil
		})
	}
	return err
}

// SwitchTag 启用或停用
func (s *TagService) SwitchTag(tagId int64, enabled int8) (err error) {
	tag := s.Mapper.TagGetById(tagId)
	if nil == tag {
		return errors.New("not found")
	}
	tag.Enabled = enabled
	tag.Modified = time.Now()
	// 事务保证一致性
	return s.Mapper.Transaction(func() error {
		if err = s.Mapper.TagSaveUp(tag); err != nil {
			return err
		}
		if err = s.refresh(tagId); err != nil {
			return err
		}
		return nil
	})
}

// refresh 刷新端口映射的开放名单
func (s *TagService) refresh(tagId int64) (err error) {
	// 标签映射：更新端口映射；获取出包含该标签的端口映射，刷新端口的白名单
	if ports := s.Mapper.PortQueryByTag([]int64{tagId}); len(ports) > 0 {
		for _, port := range ports {
			// 更新开放名单
			if err = PortWhitelist(&port, s.Mapper); err != nil {
				return err
			}
		}
	}
	return err
}

// PortWhitelist 端口映射：更新开放名单
func PortWhitelist(port *model.NatokPort, mapper dsmapper.DsMapper) (err error) {
	port.TagNilEmpty()
	port.WhitelistNilEmpty()
	tagIds := make([]int64, 0)
	for _, tag := range port.Tag {
		tagIds = append(tagIds, tag)
	}
	port.Whitelist = make([]string, 0)
	port.Tag = make([]int64, 0)
	// 标签开放名单
	if len(tagIds) > 0 {
		tags := mapper.TagFindByIds(tagIds)
		for _, tag := range tags {
			tag.WhitelistNilEmpty()
			port.Tag = append(port.Tag, tag.TagId)
			if tag.Enabled == 1 && len(tag.Whitelist) > 0 {
				port.Whitelist = append(port.Whitelist, tag.Whitelist...)
			}
		}
	}
	// 去重
	port.Whitelist = util.Distinct(port.Whitelist)
	// 端口映射：更新
	if err = mapper.PortSaveUp(port); err != nil {
		return err
	}
	// 客户端下的端口映射：更新
	if cm, ifCM := core.ClientManage.Load(port.AccessKey); ifCM {
		if pm, ifPM := cm.(*core.ClientBlocking).PortListener.Load(port.PortSign); ifPM {
			pm.(*core.PortMapping).Whitelist = port.Whitelist
		}
	}
	return err
}
