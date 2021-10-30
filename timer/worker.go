package timer

import (
	"github.com/kataras/golog"
	"go-natok-server/core"
	"go-natok-server/dsmapper"
	"time"
)

func Worker() {
	Initialization()
	//cleanExpiredNatok()
	//cleanCloseNatok()
	//HotUpdateNatok()
}
func Initialization() {
	//载入数据，绑定端口
	dsmapper.Initialization()
	core.ClientGroupManage.Range(func(key, value interface{}) bool {
		client := value.(*core.ClientBlocking)
		for _, mapping := range client.Mapping {
			core.BindPort(*mapping)
		}
		return true
	})
	//C端状态实时更新
	go func() {
		dsMapper := new(dsmapper.DsMapper)
		for {
			select {
			case chanClient := <-core.ChanClientSate:
				if dbClient := dsMapper.ClientQueryByNameOrKey(chanClient.AccessKey); dbClient != nil {
					dbClient.State = chanClient.State
					dbClient.Modified = time.Now()
					dsMapper.ClientSaveUp(dbClient)
				}
				golog.Info(chanClient)
			}
		}
	}()
}

//
////并清理掉线的连接
//func cleanCloseNatok() {
//	go func() {
//		var timeout = int64(1 * 60)        //超时时间60秒
//		var dieHandler = make([]string, 0) //断掉的连接
//		for {
//			core.NatokClientMapCache.Range(func(key, value interface{}) bool {
//				handler := value.(*core.ConnectHandler)
//				//读取时间 < 当前时间-超时时间 || 写入时间 < 当前时间-超时时间
//				if (handler.ReadTime > 0 && handler.ReadTime < (time.Now().Unix()-timeout)) ||
//					(handler.WriteTime > 0 && handler.WriteTime < (time.Now().Unix()-timeout)) {
//					handler.Conn.Close()
//					handler.MsgHandler = nil
//					handler.ConnHandler = nil
//					core.NatokClientMapCache.Delete(key)
//					dieHandler = append(dieHandler, key.(string))
//				}
//				return true
//			})
//			if len(dieHandler) > 0 {
//				for _, key := range dieHandler {
//					core.NatokClientMapCache.Delete(key)
//					log.Println("Cleaned key: ", key)
//				}
//				dieHandler = dieHandler[0:0]
//			}
//			//5秒检查一次
//			time.Sleep(time.Second * 5)
//		}
//	}()
//}
//
////HotUpdateNatok 基于数据库热更新
//func HotUpdateNatok() {
//	go func() {
//		for {
//			//TODO 客户端检查
//			clients := datasource.ClientFind(false)
//			for _, cli := range clients {
//				if cli.Enabled == 1 {
//					//客户端为启用状态，绑定它下的所有端口
//					if ports := datasource.PortGet(cli.AccessKey); ports != nil {
//						cli.UsePortNum = ports
//						core.BindPort(cli.UsePortNum()...)
//						core.BindChannelMapCache.Store(cli.AccessKey, cli)
//					}
//				} else {
//					//客户端为禁用状态，解绑它下的所有端口
//					if value, ok := core.BindChannelMapCache.Load(cli.AccessKey); ok {
//						client := value.(model.NatokClient)
//						core.UnBindPort(client.UsePortNum()...)
//					}
//					core.BindChannelMapCache.Delete(cli.AccessKey)
//				}
//				datasource.ClientUpApply(cli.ClientId)
//			}
//			//TODO 端口检查
//			ports := datasource.PortFind(false)
//			if ports != nil {
//				var portIds = make([]int64, 0)
//				for _, port := range ports {
//					if item, ok := core.BindChannelMapCache.Load(port.AccessKey); ok {
//						//已存在客户端
//						cli := item.(model.NatokClient)
//						natokPorts := cli.UsePortNum
//						for nPi, nPort := range natokPorts {
//							if port.PortNum == nPort.PortNum {
//								if port.Enabled == 1 { //添加
//									natokPorts = append(natokPorts, port)
//								} else { //删减
//									natokPorts = append(natokPorts[:nPi], natokPorts[nPi+1:]...)
//								}
//								break
//							}
//						}
//						cli.UsePortNum = append(natokPorts, port)
//						core.BindChannelMapCache.Store(port.AccessKey, cli)
//						portIds = append(portIds, port.PortId)
//					} else {
//						//需要单独查询客户端
//						if cli := datasource.GetClient(port.AccessKey); cli != nil {
//							cli.UsePortNum = append(cli.UsePortNum, port)
//							core.BindChannelMapCache.Store(port.AccessKey, &cli)
//							portIds = append(portIds, port.PortId)
//						} else {
//							continue
//						}
//					}
//					//解绑端口
//					core.UnBindPort(port.PortNum)
//					//绑定绑定端口
//					if port.Enabled == 1 {
//						core.BindPort(port.PortNum)
//					}
//				}
//				datasource.PortUpApply(portIds...)
//			}
//			//10秒检查一次
//			time.Sleep(time.Second * 10)
//		}
//	}()
//}
