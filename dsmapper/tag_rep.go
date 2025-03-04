package dsmapper

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"natok-server/model"
	"strings"
)

// 表名-NatokTag
func tableNameNatokTag() string {
	return Engine.TableName(new(model.NatokTag))
}

func (d *DsMapper) TagCountTotal(wd string) int64 {
	session := d.getSession()
	if wd != "" {
		session.Where("deleted=0 and (`tag_name` like CONCAT('%',?,'%') or `remark` like CONCAT('%',?,'%'))", wd, wd)
	} else {
		session.Where("deleted=0")
	}
	if total, err := session.Count(new(model.NatokTag)); err != nil {
		logrus.Errorf("%v", err.Error())
	} else {
		return total
	}
	return 0
}

// TagQuery 查询分页数据+总条数
func (d *DsMapper) TagQuery(wd string, page, limit int) ([]model.NatokTag, int64) {
	ret := make([]model.NatokTag, 0)
	session := d.getSession()
	if wd != "" {
		session.Where("deleted=0 and (`tag_name` like CONCAT('%',?,'%') or `remark` like CONCAT('%',?,'%'))", wd, wd)
	} else {
		session.Where("deleted=0")
	}
	session.Desc("enabled")
	session.Asc("tag_id")
	//查询分页数据
	if err := session.Limit(limit, (page-1)*limit).Find(&ret); err != nil {
		logrus.Errorf("%v", err.Error())
		return nil, 0
	}
	return ret, d.TagCountTotal(wd)
}

// TagGetByName 获取标签
func (d *DsMapper) TagGetByName(tagName string) *model.NatokTag {
	item := new(model.NatokTag)
	if ok, err := d.getSession().Where("deleted=0 and tag_name=?", tagName).Get(item); !ok {
		if err != nil {
			logrus.Errorf("%v", err.Error())
		}
		return nil
	}
	return item
}

// TagGetById 获取标签
func (d *DsMapper) TagGetById(tagId int64) *model.NatokTag {
	item := new(model.NatokTag)
	if ok, err := d.getSession().Where("deleted=0 and tag_id=?", tagId).Get(item); !ok {
		if err != nil {
			logrus.Errorf("%v", err.Error())
		}
		return nil
	}
	return item
}

// TagFindByIds 获取标签集合
func (d *DsMapper) TagFindByIds(tagIds []int64) []model.NatokTag {
	item := make([]model.NatokTag, 0)
	ids := strings.ReplaceAll(fmt.Sprintf("%v", tagIds), " ", ",")
	sub := fmt.Sprintf("tag_id in(%s)", ids[1:len(ids)-1])
	if err := d.getSession().Where("deleted=0 and " + sub).Find(&item); err != nil {
		logrus.Errorf("%v", err.Error())
		return nil
	}
	return item
}

// TagSaveUp 插入或更新标签映射
func (d *DsMapper) TagSaveUp(item *model.NatokTag) error {
	var err error = nil
	if item.TagId <= 0 {
		_, err = d.getSession().Insert(item)
	} else {
		_, err = d.getSession().
			Cols("tag_name", "remark", "whitelist", "enabled", "created", "modified", "deleted").
			Where("tag_id=?", item.TagId).Update(item)
	}
	if err != nil {
		logrus.Errorf("%v", err.Error())
	}
	return err
}
