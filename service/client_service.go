package service

import (
	"errors"
	"natok-server/core"
	"natok-server/dsmapper"
	"natok-server/model"
	"natok-server/util"
	"strings"
	"time"
)

// ClientService struct 客户端 - 服务层
type ClientService struct {
	Mapper dsmapper.DsMapper
}

// ClientQuery 列表分页
func (s *ClientService) ClientQuery(wd string, page, limit int) map[string]interface{} {
	items, total := s.Mapper.ClientQuery(wd, page, limit)
	ret := make(map[string]interface{}, 2)
	for i, item := range items {
		if cm, ifCM := core.ClientManage.Load(item.AccessKey); cm != nil && ifCM {
			item.UsePortNum = core.GetLen(&cm.(*core.ClientBlocking).PortListener)
			items[i] = item
		}
	}
	ret["items"] = items
	ret["total"] = total
	return ret
}

// ClientGet 详情
func (s *ClientService) ClientGet(clientId int64) map[string]interface{} {
	ret := make(map[string]interface{}, 2)
	ret["item"] = s.Mapper.ClientGetById(clientId)
	return ret
}

// ClientSave 保存
func (s *ClientService) ClientSave(id int64, name, accessKey string) (err error) {
	if strings.Trim(accessKey, "") == "" {
		accessKey = util.Md5(util.PreTs(util.UUID()))
	}
	// 事务保证一致性
	return s.Mapper.Transaction(func() error {
		if id <= 0 {
			//直接插入
			client := model.NatokClient{AccessKey: accessKey, ClientName: name}
			client.JoinTime = time.Now()
			client.Enabled = 1
			err = s.Mapper.ClientSaveUp(&client)
			if nil == err {
				// 启用客户端
				err = s.doClientSwitch(accessKey, 1)
			}
		} else {
			//进行更新
			client := s.Mapper.ClientGetById(id)
			if nil == client {
				return errors.New("not found")
			}
			if client.Enabled == 1 {
				return errors.New("cannot be modified")
			}
			if nil == err && client.AccessKey != accessKey {
				err = s.Mapper.PortUpdateAccessKey(client.AccessKey, accessKey)
			}
			if nil == err {
				client.ClientName = name
				client.AccessKey = accessKey
				err = s.Mapper.ClientSaveUp(client)
			}
		}
		return err
	})
}

// ClientKeys 获取所有客户端秘钥
func (s *ClientService) ClientKeys() map[string]interface{} {
	ret := make(map[string]interface{}, 2)
	ret["items"] = s.Mapper.ClientQueryKeys()
	return ret
}

// ClientSwitch 启用或停用
func (s *ClientService) ClientSwitch(clientId int64, accessKey string, enabled int8) (err error) {
	client := s.Mapper.ClientGetById(clientId)
	if nil != client && client.AccessKey == accessKey {
		client.Enabled = enabled
		client.Modified = time.Now()
		// 事务保证一致性
		return s.Mapper.Transaction(func() error {
			if err = s.Mapper.PortUpdateStateByAccessKey(accessKey, enabled); err != nil {
				return err
			}
			if err = s.Mapper.ClientSaveUp(client); err != nil {
				return err
			}
			if err = s.doClientSwitch(accessKey, 2-int(enabled)); err != nil {
				return err
			}
			return nil
		})
	}
	return errors.New("not found")
}

// ClientDelete 删除
func (s *ClientService) ClientDelete(clientId int64, accessKey string) error {
	client := s.Mapper.ClientGetById(clientId)
	if nil != client && client.AccessKey == accessKey {
		// 只能删除已停用的
		if client.Enabled == 1 {
			return errors.New("cannot be deleted")
		}
		client.Deleted = 1
		client.Enabled = 0
		client.Modified = time.Now()
		// 事务保证一致性
		return s.Mapper.Transaction(func() error {
			// 标记删除客户端映射
			if err := s.Mapper.ClientSaveUp(client); err != nil {
				return err
			}
			// 标记删除端口映射
			if err := s.Mapper.PortDeleteByAccessKey(accessKey); err != nil {
				return err
			}
			// 删除客户端连接
			core.ClientManage.Delete(accessKey)
			return nil
		})
	}
	return errors.New("not found")
}

// ClientValidate 客户端，存在验证
func (s *ClientService) ClientValidate(id int64, name string, level int32) map[string]interface{} {
	ret := make(map[string]interface{}, 2)
	if id >= 0 && name != "" && level >= 1 && level <= 2 {
		ret["state"] = dsmapper.ClientExist(id, name)
	} else {
		ret["state"] = false
	}
	return ret
}

// ClientSwitch 启停客户端
// opt=1 启用 opt=2 停用
func (s *ClientService) doClientSwitch(accessKey string, opt int) (err error) {
	switch opt {
	case 1: //启用
		cm, ifCM := core.ClientManage.Load(accessKey)
		if nil == cm || !ifCM {
			core.ClientManage.Store(accessKey, &core.ClientBlocking{
				Enabled: true, AccessKey: accessKey,
			})
		} else {
			cm.(*core.ClientBlocking).Enabled = true
		}
		if ports := s.Mapper.PortGet(accessKey); ports != nil {
			for _, port := range ports {
				port.WhitelistNilEmpty()
				mapping := &core.PortMapping{
					AccessKey: port.AccessKey, PortSign: port.PortSign,
					PortNum: port.PortNum, Intranet: port.Intranet,
					Protocol:  port.Protocol,
					Whitelist: port.Whitelist,
				}
				if err = SwitchPortMapping(mapping, 1); nil != err {
					break
				}
			}
		}
	case 2: // 停用
		if cm, ifCM := core.ClientManage.Load(accessKey); cm != nil && ifCM {
			client := cm.(*core.ClientBlocking)
			client.Enabled = false
			client.PortListener.Range(func(_, pm any) bool {
				mapping := pm.(*core.PortMapping)
				if err = SwitchPortMapping(mapping, 2); nil != err {
					return false
				}
				return true
			})
			if client.NatokHandler != nil {
				client.NatokHandler.Write(core.Message{Type: core.TypeDisabledAccessKey, Uri: accessKey})
			}
		}
	}
	return err
}
