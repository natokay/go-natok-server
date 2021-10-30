package service

import (
	"errors"
	"github.com/satori/go.uuid"
	"go-natok-server/core"
	"go-natok-server/dsmapper"
	"go-natok-server/model"
	"go-natok-server/util"
	"strings"
	"time"
)

// ClientService struct 客户端 - 服务层
type ClientService struct {
	Mapper dsmapper.DsMapper
}

// ClientQuery 查询并封装C端分页数据
func (s *ClientService) QueryClient(wd, sort string, page, limit int) map[string]interface{} {
	items, total := s.Mapper.ClientQuery(wd, sort, page, limit)
	var ret = make(map[string]interface{}, 2)
	for i, item := range items {
		if load, ok := core.ClientGroupManage.Load(item.AccessKey); load != nil && ok {
			item.UsePortNum = len(load.(*core.ClientBlocking).Mapping)
			items[i] = item
		}
	}
	ret["items"] = items
	ret["total"] = total
	return ret
}

// GetClient 获取C端信息
func (s *ClientService) GetClient(clientId int64) map[string]interface{} {
	var ret = make(map[string]interface{}, 2)
	ret["item"] = s.Mapper.ClientGetById(clientId)
	return ret
}

// SaveClient 保存C端
func (s *ClientService) SaveClient(id int64, name, accessKey string) error {
	//TODO 事务保证一致性
	err := s.Mapper.Transaction()
	if nil != err {
		return err
	}
	if strings.Trim(accessKey, "") == "" {
		accessKey = util.Md5(time.Now().String() + uuid.NewV4().String())
	}
	if id <= 0 {
		//直接插入
		client := model.NatokClient{AccessKey: accessKey, ClientName: name, JoinTime: time.Now(), Enabled: 1}
		err = s.Mapper.ClientSaveUp(&client)
	} else {
		//进行更新
		if client := s.Mapper.ClientGetById(id); client != nil {
			oldAccessKey := client.AccessKey
			client.ClientName = name
			client.Apply = 0
			client.AccessKey = accessKey

			err := s.Mapper.ClientSaveUp(client)
			if nil == err {
				err = dsmapper.PortUpdateAccessKey(oldAccessKey, accessKey)
			}
			if nil == err {
				if accessKey != oldAccessKey {
					err = SwitchClientBlocking(oldAccessKey, 3)
				} else {
					err = SwitchClientBlocking(oldAccessKey, 2)
				}
			}
		}
	}
	if nil == err {
		err = SwitchClientBlocking(accessKey, 1)
	}
	if nil == err {
		s.Mapper.Commit()
	} else {
		s.Mapper.Rollback()
	}
	return err
}

// ClientKeys 获取所有C端秘钥
func (s *ClientService) ClientKeys() map[string]interface{} {
	var ret = make(map[string]interface{}, 2)
	ret["items"] = s.Mapper.ClientQueryKeys()
	return ret
}

// SwitchClient 启用或停用C端
func (s *ClientService) SwitchClient(clientId int64, accessKey string, enabled int8) error {
	client := s.Mapper.ClientGetById(clientId)
	if nil != client && client.AccessKey == accessKey {
		client.Apply = 1
		client.Enabled = enabled
		client.Modified = time.Now()

		//TODO 事务保证一致性
		err := s.Mapper.Transaction()
		if nil == err {
			err = s.Mapper.PortUpdateStateByAccessKey(accessKey, enabled)
		}
		if nil == err {
			err = s.Mapper.ClientSaveUp(client)
		}
		if nil == err {
			err = SwitchClientBlocking(accessKey, 2-int(enabled))
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

// DeleteClient 标记删除C端
func (s *ClientService) DeleteClient(clientId int64, accessKey string) error {
	client := s.Mapper.ClientGetById(clientId)
	if nil != client && client.AccessKey == accessKey {
		client.Deleted = 1
		client.Apply = 1
		client.Enabled = 0
		client.Modified = time.Now()

		//TODO 事务保证一致性
		err := s.Mapper.Transaction()
		if nil == err {
			err = s.Mapper.ClientSaveUp(client)
		}
		if nil == err {
			err = s.Mapper.PortUpdateStateByAccessKey(accessKey, 0)
		}
		if nil == err {
			err = SwitchClientBlocking(accessKey, 3)
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

// ValidateClient C端，存在验证
func (s *ClientService) ValidateClient(id int64, name string, level int32) map[string]interface{} {
	var ret = make(map[string]interface{}, 2)
	if id >= 0 && name != "" && level >= 1 && level <= 2 {
		ret["state"] = dsmapper.ClientExist(id, name)
	} else {
		ret["state"] = false
	}
	return ret
}

//SwitchClientBlocking 启停C端
// opt=1 启用 opt=2 停用 opt=3 停用&&删除
func SwitchClientBlocking(accessKey string, opt int) (err error) {
	switch opt {
	case 1: //启用
		clientBlocking, ok := core.ClientGroupManage.Load(accessKey)
		if nil == clientBlocking || !ok {
			core.ClientGroupManage.Store(accessKey, &core.ClientBlocking{
				Enabled: true, AccessKey: accessKey, Mapping: make(map[string]*core.PortMapping, 0),
			})
		} else {
			clientBlocking.(*core.ClientBlocking).Enabled = true
		}
		if ports := dsmapper.PortGet(accessKey); ports != nil {
			for _, port := range ports {
				if nil == err {
					err = SwitchPortMapping(core.PortMapping{
						AccessKey: port.AccessKey, Sign: port.Sign,
						Port: port.PortNum, Intranet: port.Intranet,
						Domain: port.Domain, Protocol: port.Protocol,
					}, 1)
				} else {
					break
				}
			}
		}
	case 2: // 停用
		if load, ok := core.ClientGroupManage.Load(accessKey); load != nil && ok {
			clientBlocking := load.(*core.ClientBlocking)
			clientBlocking.Enabled = false
			if nil != clientBlocking.Listener {
				err = clientBlocking.Listener.Close()
			}
			portMappingMap := clientBlocking.Mapping
			if nil != portMappingMap {
				for _, mapping := range portMappingMap {
					if nil == err {
						err = SwitchPortMapping(*mapping, 2)
					} else {
						break
					}
				}
			}
		}
	case 3: // 停用 && 删除
		if err = SwitchClientBlocking(accessKey, 2); err == nil {
			core.ClientGroupManage.Delete(accessKey)
		}
	}
	return err
}
