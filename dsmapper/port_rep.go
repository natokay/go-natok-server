package dsmapper

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/kataras/golog"
	"natok-server/model"
	"strconv"
	"strings"
	"time"
)

// 表明-NatokPort
func tableNameNatokPort() string {
	return Engine.TableName(new(model.NatokPort))
}

// PortFind 查询端口
func PortFind(isAll bool) (ret []model.NatokPort) {
	sql := "1=1"
	if isAll {
		sql += " and enabled=1"
	} else {
		sql += " and apply=0"
	}
	if err := Engine.Where(sql).Find(&ret); err != nil {
		golog.Error(err)
		return nil
	}
	return ret
}

// PortGet 获取端口
func PortGet(key string) (ret []model.NatokPort) {
	if err := Engine.Where("enabled=1 and access_key=?", key).Find(&ret); err != nil {
		golog.Error(err)
		return nil
	}
	return ret
}

// PortUpApply 更新端口数据为已应用
func PortUpApply(ids ...int64) {
	if len(ids) == 0 {
		return
	}
	sql, args, err := sqlx.In("update "+tableNameNatokPort()+" set apply=1 where port_id in (?)", ids)
	if err != nil {
		golog.Error(err)
		return
	}
	var sar []interface{}
	sar = append(append(sar, sql), args...)
	if _, err := Engine.Exec(sar...); err != nil {
		golog.Error(err)
		return
	}
}

////////////////////////////////////////////////////////////////////

// PortQuery 查询分页数据+总条数
func (d *DsMapper) PortQuery(wd, sort string, page, limit int) ([]model.NatokPort, int64) {
	ret := make([]model.NatokPort, 0)
	session := d.getSession()
	if wd != "" {
		session.Where("deleted=0 and access_key=?", wd)
	} else {
		session.Where("deleted=0")
	}
	//查询分页数据
	if err := session.Limit(limit, (page-1)*limit).Find(&ret); err != nil {
		golog.Error(err)
		return nil, 0
	}
	return ret, d.PortCountTotal()
}

// PortGetById 获取端口
func (d *DsMapper) PortGetById(portId int64) *model.NatokPort {
	item := new(model.NatokPort)
	if ok, err := d.getSession().Where("port_id=?", portId).Get(item); !ok {
		if err != nil {
			golog.Error(err)
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
			Cols("port_num", "intranet", "domain", "protocol", "expire_at", "remark", "enabled", "state", "apply", "modified", "deleted").
			Where("port_id=?", item.PortId).Update(item)
	}
	if err != nil {
		golog.Error(err)
	}
	return err
}

// PortUpdateAccessKey 更新AccessKey
func PortUpdateAccessKey(oldAccessKey, newAccessKey string) error {
	item := model.NatokPort{
		AccessKey: newAccessKey, Apply: 0, Modified: time.Now(),
	}
	_, err := Engine.Where("access_key=?", oldAccessKey).Update(&item)
	if err != nil {
		golog.Error(err)
	}
	return err
}

// PortUpdateStateByAccessKey 根据AccessKey更新Sate
func (d *DsMapper) PortUpdateStateByAccessKey(accessKey string, state int8) error {
	item := model.NatokPort{State: state, Modified: time.Now()}
	_, err := d.getSession().Cols("state", "modified").Where("access_key=?", accessKey).Update(&item)
	if err != nil {
		golog.Error(err)
	}
	return err
}

// PortExist 端口映射，是否存在于其他
func (d *DsMapper) PortExist(portId int64, value string) bool {
	exist, err := d.getSession().Where("port_id!=? and (port_num=? or domain=?)", portId, value, value).Exist(new(model.NatokPort))
	if err != nil {
		golog.Error(err)
	}
	return exist
}

// PortGroupByProtocol 端口协议统计
func (d *DsMapper) PortGroupByProtocol() map[string]interface{} {
	ret := make(map[string]interface{}, 0)
	sql := "select protocol,count(protocol) as count from " + tableNameNatokPort() + " where deleted=0 group by protocol"
	if result, err := d.getSession().Query(sql); err != nil {
		golog.Error(err)
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
	sql := "select count(1) as count from " + tableNameNatokPort() + " where expire_at %s now()"
	if expired {
		sql = fmt.Sprintf(sql, "<")
	} else {
		sql = fmt.Sprintf(sql, ">=")
	}
	if result, err := d.getSession().Query(sql); err != nil {
		golog.Error(err)
	} else {
		count, _ := strconv.Atoi(string(result[0]["count"]))
		return count
	}
	return 0
}

// PortCountTotal 端口总量统计
func (d *DsMapper) PortCountTotal() int64 {
	if total, err := d.getSession().Where("deleted=0").Count(new(model.NatokPort)); err != nil {
		golog.Error(err)
	} else {
		return total
	}
	return 0
}
