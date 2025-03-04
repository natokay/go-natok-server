package timer

import (
	"github.com/sirupsen/logrus"
	"natok-server/core"
	"natok-server/dsmapper"
	"natok-server/support"
	"natok-server/util"
	"time"
)

func Worker() {
	Initialization()
}
func Initialization() {
	dsMapper := new(dsmapper.DsMapper)
	//载入数据，绑定端口
	dsMapper.StateRest()
	core.ClientManage.Range(func(key, value interface{}) bool {
		client := value.(*core.ClientBlocking)
		client.PortListener.Range(func(_, pm any) bool {
			mapping := pm.(*core.PortMapping)
			if err := core.BindPort(mapping); err != nil {
				logrus.Errorf("%v", err.Error())
			}
			return true
		})
		return true
	})
	//客户端状态实时更新
	go func() {
		for {
			select {
			case chanClient := <-core.ChanClientSate:
				if dbClient := dsMapper.ClientQueryByNameOrKey(chanClient.AccessKey); dbClient != nil {
					dbClient.State = chanClient.State
					dbClient.Modified = time.Now()
					_ = dsMapper.ClientSaveUp(dbClient)
				}
				logrus.Infof("Client %s is onlie", chanClient.AccessKey)
			}
		}
	}()
	// 自动停用已过期端口
	go func() {
		for {
			select {
			case <-time.After(time.Minute * 15):
				if ports := dsMapper.PortGetExpired(); len(ports) > 0 {
					for _, port := range ports {
						if cm, ifCM := core.ClientManage.Load(port.AccessKey); cm != nil && ifCM {
							client := cm.(*core.ClientBlocking)
							if pm, ifPM := client.PortListener.Load(port.PortSign); pm != nil && ifPM {
								mapping := pm.(*core.PortMapping)
								_ = core.UnBindPort(mapping)
							}
						}
					}
					dsMapper.DisableExpiredPort()
				}
			}
		}
	}()
	// 定时清理空闲连接
	go func() {
		for {
			select {
			case <-time.After(time.Second * 5):
				now := time.Now()
				idleTimeout := time.Duration(support.AppConf.Natok.Server.ChanPool.IdleTimeout) * time.Second
				core.ConnectManage.Range(func(accessKey, cn any) bool {
					blocking := cn.(*core.ConnectBlocking)
					blocking.PortSignMap.Range(func(sign, ids any) bool {
						portSign, accessIds := sign.(string), ids.([]string)
						for _, accessId := range accessIds {
							if connect, ifConnect := blocking.AccessIdMap.Load(accessId); ifConnect && connect != nil {
								extra := connect.(*core.ExtraConnectHandler)
								// 非活跃状态
								if extra.ConnHandler.Active == false {
									// UDP无连接的协议，不需要频繁清理
									if extra.Protocol == core.Udp && now.Sub(extra.ConnHandler.WriteTime) > idleTimeout {
										blocking.AccessIdMap.Delete(accessId)
										blocking.PortSignMap.Store(portSign, util.RemoveIf(accessIds, accessId))
										extra.ConnHandler.MsgHandler.Close(extra.ConnHandler)
									}
									if extra.Protocol == core.Tcp {
										blocking.AccessIdMap.Delete(accessId)
										blocking.PortSignMap.Store(portSign, util.RemoveIf(accessIds, accessId))
									}
									logrus.Debugf("Clean idle connect %s", accessId)
								}
							}
						}
						return true
					})
					return true
				})
			}
		}
	}()
}
