package dsmapper

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"natok-server/model"
	"natok-server/support"
	"strconv"
	"strings"
	"time"
)

// 表名-NatokPort
func tableNameNatokPort() string {
	return Engine.TableName(new(model.NatokPort))
}

// DisableExpiredPort 停用已过期端口
func (d *DsMapper) DisableExpiredPort() {
	sql, args, err := sqlx.In("update " + tableNameNatokPort() + " set enabled=0,state=1,apply=1 where expire_at <= CURRENT_TIMESTAMP")
	if err != nil {
		logrus.Errorf("%v", err.Error())
		return
	}
	var sar []interface{}
	sar = append(append(sar, sql), args...)
	if _, err := d.getSession().Exec(sar...); err != nil {
		logrus.Errorf("%v", err.Error())
		return
	}
}

// PortFind 查询端口
func (d *DsMapper) PortFind(isAll bool) (ports []model.NatokPort) {
	sql := "deleted=0 and expire_at > CURRENT_TIMESTAMP"
	if isAll {
		sql += " and enabled=1"
	} else {
		sql += " and apply=0"
	}
	if err := d.getSession().Where(sql).Find(&ports); err != nil {
		logrus.Errorf("%v", err.Error())
		return nil
	}
	return ports
}

// PortGetExpired 获取已过期的端口
func (d *DsMapper) PortGetExpired() (ports []model.NatokPort) {
	if err := d.getSession().Where("deleted=0 and enabled=1 and expire_at <= CURRENT_TIMESTAMP").Find(&ports); err != nil {
		logrus.Errorf("%v", err.Error())
		return nil
	}
	return ports
}

// PortGet 获取端口
func (d *DsMapper) PortGet(key string) (ret []model.NatokPort) {
	if err := d.getSession().Where("deleted=0 and enabled=1 and access_key=?", key).Find(&ret); err != nil {
		logrus.Errorf("%v", err.Error())
		return nil
	}
	return ret
}

// PortUpApply 更新端口数据为已应用
func (d *DsMapper) PortUpApply(ids ...int64) {
	if len(ids) == 0 {
		return
	}
	sql, args, err := sqlx.In("update "+tableNameNatokPort()+" set apply=1 where port_id in (?)", ids)
	if err != nil {
		logrus.Errorf("%v", err.Error())
		return
	}
	var sar []interface{}
	sar = append(append(sar, sql), args...)
	if _, err := d.getSession().Exec(sar...); err != nil {
		logrus.Errorf("%v", err.Error())
		return
	}
}

////////////////////////////////////////////////////////////////////

// PortQuery 查询分页数据+总条数
func (d *DsMapper) PortQuery(wd string, page, limit int) ([]model.NatokPort, int64) {
	ret := make([]model.NatokPort, 0)
	session := d.getSession()
	if wd != "" {
		session.Where("deleted=0 and access_key=?", wd)
	} else {
		session.Where("deleted=0")
	}
	session.Desc("enabled")
	session.Asc("port_id")
	//查询分页数据
	if err := session.Limit(limit, (page-1)*limit).Find(&ret); err != nil {
		logrus.Errorf("%v", err.Error())
		return nil, 0
	}
	return ret, d.PortCountTotal(wd)
}

// PortQueryByTag 根据标签查询端口
func (d *DsMapper) PortQueryByTag(tagIds []int64) []model.NatokPort {
	ret := make([]model.NatokPort, 0)
	session := d.getSession()
	// mysql数据库
	mySql := func() string {
		sql := ""
		for i, v := range tagIds {
			if i > 0 {
				sql += " or "
			}
			sql += fmt.Sprintf("JSON_CONTAINS(`tag`, '%d')", v)
		}
		return fmt.Sprintf(" and (%s)", sql)
	}
	// sqlite数据库
	sqlLite := func() string {
		ids := strings.ReplaceAll(fmt.Sprintf("%v", tagIds), " ", ",")
		sub := fmt.Sprintf("(select 1 from json_each(`tag`) where json_each.value in (%s))", ids[1:len(ids)-1])
		return fmt.Sprintf(" and exists %s", sub)
	}
	sub := sqlLite()
	if support.AppConf.Natok.Db.Type == "mysql" {
		sub = mySql()
	}
	if err := session.Where("deleted=0 " + sub).Find(&ret); err != nil {
		logrus.Errorf("%v", err.Error())
	}
	return ret
}

// PortGetById 获取端口
func (d *DsMapper) PortGetById(portId int64) *model.NatokPort {
	item := new(model.NatokPort)
	if ok, err := d.getSession().Where("deleted=0 and port_id=?", portId).Get(item); !ok {
		if err != nil {
			logrus.Errorf("%v", err.Error())
		}
		return nil
	}
	return item
}

// PortSaveUp 插入或更新端口映射
func (d *DsMapper) PortSaveUp(item *model.NatokPort) error {
	var err error = nil
	if item.PortId <= 0 {
		_, err = d.getSession().Insert(item)
	} else {
		_, err = d.getSession().
			Cols("port_scope", "port_num", "intranet", "protocol", "expire_at", "whitelist", "tag", "remark", "enabled", "state", "apply", "modified", "deleted").
			Where("port_id=?", item.PortId).Update(item)
	}
	if err != nil {
		logrus.Errorf("%v", err.Error())
	}
	return err
}

// PortUpdateAccessKey 更新AccessKey
func (d *DsMapper) PortUpdateAccessKey(oldAccessKey, newAccessKey string) error {
	item := model.NatokPort{
		AccessKey: newAccessKey, Apply: 0, Modified: time.Now(),
	}
	_, err := d.getSession().Where("access_key=?", oldAccessKey).Update(&item)
	if err != nil {
		logrus.Errorf("%v", err.Error())
	}
	return err
}

// PortUpdateStateByAccessKey 根据AccessKey更新Sate
func (d *DsMapper) PortUpdateStateByAccessKey(accessKey string, state int8) error {
	item := model.NatokPort{State: state, Modified: time.Now()}
	_, err := d.getSession().Cols("state", "modified").Where("access_key=?", accessKey).Update(&item)
	if err != nil {
		logrus.Errorf("%v", err.Error())
	}
	return err
}

// PortDeleteByAccessKey 删除端口映射
func (d *DsMapper) PortDeleteByAccessKey(accessKey string) error {
	item := model.NatokPort{Deleted: 1, Enabled: 0, State: 0, Modified: time.Now()}
	_, err := d.getSession().Cols("state", "modified").Where("access_key=?", accessKey).Update(&item)
	if err != nil {
		logrus.Errorf("%v", err.Error())
	}
	return err
}

// PortExist 端口映射，是否存在于其他
func (d *DsMapper) PortExist(portId int64, portNum int, protocol string) bool {
	exist, err := d.getSession().Where("deleted=0 and port_id!=? and port_num=? and protocol=?", portId, portNum, protocol).Exist(new(model.NatokPort))
	if err != nil {
		logrus.Errorf("%v", err.Error())
	}
	return exist
}

// PortGroupByProtocol 端口协议统计
func (d *DsMapper) PortGroupByProtocol() map[string]interface{} {
	ret := make(map[string]interface{}, 0)
	sql := "select protocol,count(protocol) as count from " + tableNameNatokPort() + " where deleted=0 group by protocol"
	if result, err := d.getSession().Query(sql); err != nil {
		logrus.Errorf("%v", err.Error())
	} else {
		for _, m := range result {
			protocol := strings.ToUpper(string(m["protocol"]))
			count, _ := strconv.Atoi(string(m["count"]))
			ret[protocol] = count
		}
	}
	return ret
}

// PortExpiredCount 端口过期统计
// expired true = expired count, false = not expired count
func (d *DsMapper) PortExpiredCount(expired bool) int {
	sql := "select count(1) as count from " + tableNameNatokPort() + " where deleted=0 and expire_at %s CURRENT_TIMESTAMP"
	if expired {
		sql = fmt.Sprintf(sql, "<")
	} else {
		sql = fmt.Sprintf(sql, ">=")
	}
	if result, err := d.getSession().Query(sql); err != nil {
		logrus.Errorf("%v", err.Error())
	} else {
		count, _ := strconv.Atoi(string(result[0]["count"]))
		return count
	}
	return 0
}

// PortCountTotal 端口总量统计
func (d *DsMapper) PortCountTotal(accessKey string) int64 {
	where := d.getSession().Where("deleted=0")
	if accessKey != "" {
		where.And("access_key=?", accessKey)
	}
	if total, err := where.Count(new(model.NatokPort)); err != nil {
		logrus.Errorf("%v", err.Error())
	} else {
		return total
	}
	return 0
}
