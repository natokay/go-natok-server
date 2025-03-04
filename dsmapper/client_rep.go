package dsmapper

import (
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"natok-server/model"
	"strconv"
)

// 表名-NatokClient
func tableNameNatokClient() string {
	return Engine.TableName(new(model.NatokClient))
}

// ClientFindAll 查询客户端
func (d *DsMapper) ClientFindAll() (ret []model.NatokClient) {
	if err := d.getSession().Where("deleted=0").Find(&ret); err != nil {
		logrus.Errorf("%v", err.Error())
		return nil
	}
	return ret
}

// //////////////////////////////////////////////////////////////////
// ClientQueryByNameOrKey 查询客户端根据 kw
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
			logrus.Errorf("%v", err.Error())
		}
		return nil
	}
	return item
}

// ClientQuery 查询分页数据+总条数
func (d *DsMapper) ClientQuery(wd string, page, limit int) (ret []model.NatokClient, total int64) {
	var err error
	session := d.getSession()
	for i := 0; i <= 1; i++ {
		if wd != "" {
			session.Where("deleted=0 and (client_id=? or client_name like CONCAT('%',?,'%') or access_key=?)", wd, wd, wd)
		} else {
			session.Where("deleted=0")
		}
		session.Desc("enabled", "state")
		session.Asc("client_id")
		switch i {
		case 0: //查询分页数据
			if err = session.Limit(limit, (page-1)*limit).Find(&ret); err != nil {
				logrus.Errorf("%v", err.Error())
				return nil, total
			}
		case 1: //查询总条数
			if total, err = session.Count(new(model.NatokClient)); err != nil {
				logrus.Errorf("%v", err.Error())
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
			logrus.Errorf("%v", err.Error())
		}
		return nil
	}
	return ret
}

// ClientStateReset 重置客户端状态
func (d *DsMapper) ClientStateReset() {
	update, err := d.getSession().Cols("state").Where("state=?", 1).Update(new(model.NatokClient))
	if nil != err {
		logrus.Errorf("%v", err.Error())
	}
	if update > 0 {
		logrus.Info("Reset client state ", update, " row")
	}
}

// ClientSaveUp 插入或更新客户端
func (d *DsMapper) ClientSaveUp(item *model.NatokClient) error {
	var err error = nil
	if item.ClientId <= 0 {
		_, err = d.getSession().Insert(item)
	} else {
		_, err = d.getSession().Cols("client_name", "access_key", "enabled", "state", "modified", "deleted").
			Where("client_id=?", item.ClientId).Update(item)
	}
	if err != nil {
		logrus.Errorf("%v", err.Error())
	}
	return err
}

// ClientQueryKeys 获取全部客户端密钥
func (d *DsMapper) ClientQueryKeys() (ret []model.NatokClient) {
	if err := d.getSession().Cols("client_name", "access_key", "enabled").Where("deleted=0").Find(&ret); err != nil {
		logrus.Errorf("%v", err.Error())
	}
	return ret
}

// ClientExist 客户端，是否存在于其他
func ClientExist(clientId int64, value string) bool {
	exist, err := Engine.Where("deleted=0 and client_id !=? and (client_name=? or access_key=?)", clientId, value, value).Exist(new(model.NatokClient))
	if err != nil {
		logrus.Errorf("%v", err.Error())
	}
	return exist
}

// ClientGroupByState 客户端在线状态
func (d *DsMapper) ClientGroupByState() map[string]interface{} {
	ret := make(map[string]interface{}, 0)
	sql := "select state,count(state) as count from " + tableNameNatokClient() + " where deleted=0 group by state"
	if result, err := d.getSession().Query(sql); err != nil {
		logrus.Errorf("%v", err.Error())
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

// ClientGroupByEnabled 客户端启用
func (d *DsMapper) ClientGroupByEnabled() map[string]interface{} {
	ret := make(map[string]interface{}, 0)
	sql := "select enabled,count(state) as count from " + tableNameNatokClient() + " where deleted=0 group by enabled"
	if result, err := d.getSession().Query(sql); err != nil {
		logrus.Errorf("%v", err.Error())
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
