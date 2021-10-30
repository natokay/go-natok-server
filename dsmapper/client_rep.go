package dsmapper

import (
	"github.com/go-xorm/xorm"
	"github.com/jmoiron/sqlx"
	"github.com/kataras/golog"
	"go-natok-server/model"
	"strconv"
)

// 表明-NatokClient
func tableNameNatokClient() string {
	return Engine.TableName(new(model.NatokClient))
}

// ClientFind 查询客户端
func ClientFind(isAll bool) (ret []model.NatokClient) {
	sql := "select * from " + tableNameNatokClient() + " where deleted=0"
	if !isAll {
		sql += " and apply=0"
	}
	if err := Engine.SQL(sql).Find(&ret); err != nil {
		golog.Error(err)
		return nil
	}
	return ret
}

// ClientUpApply 更新客户端数据为已应用
func ClientUpApply(ids ...int64) {
	sql, args, err := sqlx.In("update "+tableNameNatokClient()+" set apply=1 where client_id in (?)", ids)
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
// ClientQueryByNameOrKey 查询C端根据 kw
func (d *DsMapper) ClientQueryByNameOrKey(kw string) *model.NatokClient {
	var session *xorm.Session
	if kw != "" {
		session = d.getSession().Where("deleted=0 and(client_name like CONCAT('%',?,'%') or access_key=?)", kw, kw)
	} else {
		session = d.getSession()
	}
	item := new(model.NatokClient)
	if ok, err := session.Get(item); !ok {
		if err != nil {
			golog.Error(err)
		}
		return nil
	}
	return item
}

// ClientQuery 查询分页数据+总条数
func (d *DsMapper) ClientQuery(wd, sort string, page, limit int) (ret []model.NatokClient, total int64) {
	var err error
	session := d.getSession()
	for i := 0; i <= 1; i++ {
		if wd != "" {
			session.Where("deleted=0 and (client_id=? or name like CONCAT('%',?,'%') or access_key=?)", wd, wd, wd)
		} else {
			session.Where("deleted=0")
		}
		switch i {
		case 0: //查询分页数据
			if err = session.Limit(limit, (page-1)*limit).Find(&ret); err != nil {
				golog.Error(err)
				return nil, total
			}
		case 1: //查询总条数
			if total, err = session.Count(new(model.NatokClient)); err != nil {
				golog.Error(err)
				return nil, total
			}
		}
	}
	return ret, total
}

// ClientGetById 获取客户端
func (d *DsMapper) ClientGetById(clientId int64) *model.NatokClient {
	ret := new(model.NatokClient)
	if ok, err := d.getSession().Where("deleted=0 and client_id=?", clientId).Get(ret); !ok {
		if err != nil {
			golog.Error(err)
		}
		return nil
	}
	return ret
}

// ClientResetState 重置C端状态
func ClientResetState() {
	update, err := Engine.Cols("state").Where("state=?", 1).Update(new(model.NatokClient))
	if nil != err {
		golog.Error(err)
	}
	if update > 0 {
		golog.Info("Reset client state ", update, " row")
	}
}

// ClientSaveUp 插入或更新客户端
func (d *DsMapper) ClientSaveUp(item *model.NatokClient) error {
	var err error = nil
	if item.ClientId <= 0 {
		_, err = d.getSession().Insert(item)
	} else {
		_, err = d.getSession().Cols("client_name", "access_key", "enabled", "state", "apply", "modified", "deleted").
			Where("client_id=?", item.ClientId).Update(item)
	}
	if err != nil {
		golog.Error(err)
	}
	return err
}

// ClientQueryKeys 获取全部C端密钥
func (d *DsMapper) ClientQueryKeys() (ret []model.NatokClient) {
	if err := d.getSession().Cols("client_name", "access_key", "enabled").Where("deleted=0").Find(&ret); err != nil {
		golog.Error(err)
	}
	return ret
}

// ClientExist C端，是否存在于其他
func ClientExist(clientId int64, value string) bool {
	exist, err := Engine.Where("client_id !=? and (client_name=? or access_key=?)", clientId, value, value).Exist(new(model.NatokClient))
	if err != nil {
		golog.Error(err)
	}
	return exist
}

// ClientGroupByState C端在线状态
func (d *DsMapper) ClientGroupByState() map[string]interface{} {
	ret := make(map[string]interface{}, 0)
	sql := "select state,count(state) as count from " + tableNameNatokClient() + " where deleted=0 group by state"
	if result, err := d.getSession().Query(sql); err != nil {
		golog.Error(err)
	} else {
		for _, m := range result {
			count, _ := strconv.Atoi(string(m["count"]))
			switch string(m["state"]) {
			case "0":
				ret["离线"] = count
			case "1":
				ret["在线"] = count
			}
		}
	}
	return ret
}

// ClientGroupByEnabled C端启用
func (d *DsMapper) ClientGroupByEnabled() map[string]interface{} {
	ret := make(map[string]interface{}, 0)
	sql := "select enabled,count(state) as count from " + tableNameNatokClient() + " where deleted=0 group by enabled"
	if result, err := d.getSession().Query(sql); err != nil {
		golog.Error(err)
	} else {
		for _, m := range result {
			count, _ := strconv.Atoi(string(m["count"]))
			switch string(m["enabled"]) {
			case "0":
				ret["停用"] = count
			case "1":
				ret["启用"] = count
			}
		}
	}
	return ret
}
