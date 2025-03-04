package core

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"natok-server/util"
	"net"
	"strings"
	"time"
)

// ExtraConnectHandler struct 外部端口请求服务处理
type ExtraConnectHandler struct {
	AccessId    string    //接受连接的ID
	AccessKey   string    //绑定的客户端秘钥
	Sign        string    //映射签名
	Protocol    string    //传输协议
	Port        int       //绑定的端口
	activated   bool      //已经激活
	ChanActive  chan bool //激活通道
	ConnHandler *ConnectHandler
}

// Encode 编码消息
func (e *ExtraConnectHandler) Encode(msg interface{}) []byte {
	if msg == nil {
		return []byte{}
	}
	return msg.([]byte)
}

// Decode 解码消息
func (e *ExtraConnectHandler) Decode(buf []byte) (interface{}, int) {
	return buf, len(buf)
}

// Close 关闭处理
func (e *ExtraConnectHandler) Close(handler *ConnectHandler) {
	// 通知客户端关闭连接
	if natokHandler := handler.ConnHandler; natokHandler != nil {
		natokHandler.Write(Message{Type: TypeDisconnect, Serial: natokHandler.MsgHandler.(*NatokConnectHandler).Serial, Uri: e.AccessId})
		handler.ConnHandler = nil
	}
	// 关闭连接
	if handler.Conn != nil {
		if conn := handler.Conn.Conn; conn != nil {
			_ = conn.Close()
		}
		if packet := handler.Conn.Packet; packet != nil {
			packet.Addr = nil
		}
	}
	handler.Active = false
}

// Receive 请求接收
func (e *ExtraConnectHandler) Receive(handler *ConnectHandler, data interface{}) {
	// 判断客户端是否已连接，未连接则等待连接
	if e.activated == false {
		select {
		case active := <-e.ChanActive:
			if active {
				e.activated = true
				break
			}
		case <-time.After(time.Second * 15):
			logrus.Errorf("Receive AccessKey %s -> %s, PortNum %d wait 15 second, connection timeout. ", e.AccessKey, e.AccessId, e.Port)
			e.Close(handler)
			return
		}
	}
	// 将请求转发到客户端
	if natokHandler := handler.ConnHandler; natokHandler != nil {
		msg := Message{Type: TypeTransfer, Serial: natokHandler.MsgHandler.(*NatokConnectHandler).Serial, Uri: e.AccessId, Data: data.([]byte)}
		natokHandler.Write(msg)
	}
}

// Activate 激活（将内部网地址发送给客户端建立连接）
func (e *ExtraConnectHandler) Activate() {
	// 若外部端口已完成绑定，激活该请求
	ClientManage.Range(func(_, cm any) bool {
		ifBool := true
		client := cm.(*ClientBlocking)
		client.PortListener.Range(func(sign, pm any) bool {
			portMapping := pm.(*PortMapping)
			if e.Port == portMapping.PortNum {
				if nil != client.NatokHandler {
					if cn, ifCN := ConnectManage.Load(e.AccessKey); cn != nil && ifCN {
						blocking := cn.(*ConnectBlocking)
						if connect, ifConnect := blocking.AccessIdMap.Load(e.AccessId); connect != nil && ifConnect {
							// 已存在连接通道
							ifBool = false
							return ifBool
						}
						// 追加访问通道
						if signs, ifSign := blocking.PortSignMap.Load(e.Sign); signs != nil && ifSign {
							blocking.PortSignMap.Store(e.Sign, append(signs.([]string), e.AccessId))
						} else {
							blocking.PortSignMap.Store(e.Sign, []string{e.AccessId})
						}
						blocking.AccessIdMap.Store(e.AccessId, e)
					} else {
						// 创建端口连接通道
						blocking := &ConnectBlocking{}
						blocking.PortSignMap.Store(e.Sign, []string{e.AccessId})
						blocking.AccessIdMap.Store(e.AccessId, e)
						ConnectManage.Store(e.AccessKey, blocking)
					}
					// 激活连接通道
					e.Protocol = portMapping.Protocol
					if e.Protocol != Udp {
						e.Protocol = Tcp
					}
					msg := Message{Type: TypeConnectIntra, Serial: e.ConnHandler.ConnHandler.MsgHandler.(*NatokConnectHandler).Serial, Net: e.Protocol, Uri: e.AccessId, Data: []byte(portMapping.Intranet)}
					e.ConnHandler.ConnHandler.Write(msg)
					ifBool = false
					return ifBool
				}
			}
			return ifBool
		})
		return ifBool
	})
}

func getExtra(accessKey, accessId string, fn func() *ExtraConnectHandler) *ExtraConnectHandler {
	if cn, ifCN := ConnectManage.Load(accessKey); cn != nil && ifCN {
		blocking := cn.(*ConnectBlocking)
		if connect, ifConnect := blocking.AccessIdMap.Load(accessId); connect != nil && ifConnect {
			return connect.(*ExtraConnectHandler)
		}
	}
	return fn()
}

func udpListen(mapping *PortMapping) error {
	// 绑定服务端的端口
	packet, err := net.ListenPacket(Udp, util.ToAddress(mapping.PortScope, mapping.PortNum))
	if err != nil {
		logrus.Errorf("Listen udp port %d failed! %v", mapping.PortNum, err.Error())
		return err
	}
	logrus.Infof("Bind udp listen %s", packet.LocalAddr())
	mapping.Listener = &DualListener{PacketConn: packet}
	go func(packet net.PacketConn) {
		for {
			if mapping.Enabled == false {
				logrus.Infof("Port %d not enabled, exit udp listen %s", mapping.PortNum, packet.LocalAddr())
				_ = packet.Close()
				break
			}
			// 接收数据 udp max packet 64kb
			buf := make([]byte, 1024*64)
			n, addr, err := packet.ReadFrom(buf)
			if err != nil || n == 0 {
				logrus.Errorf("%v", err.Error())
				break
			}
			if cm, ifCM := ClientManage.Load(mapping.AccessKey); cm != nil && ifCM {
				// 客户端: 启用 && 在线
				client := cm.(*ClientBlocking)
				if !client.Enabled || client.NatokHandler == nil || !client.NatokHandler.Active {
					_ = packet.Close()
					break
				}
				// 限制开放名单
				pm, _ := client.PortListener.Load(mapping.PortSign)
				if nil == pm || util.NoneMatch(pm.(*PortMapping).Whitelist, func(ip string) bool {
					return strings.Contains(addr.String(), ip)
				}, false) {
					continue
				}
				sprintf := fmt.Sprintf("Accept connect udp://%s from %s %d", addr.String(), mapping.Protocol, mapping.PortNum)
				logrus.Debugf("%s", sprintf)
				// 将外部连接与内部连接关联
				extra := getExtra(mapping.AccessKey, Udp+Protocol+addr.String(), func() *ExtraConnectHandler {
					// 获取内部连接通道
					natokHandler := client.NatokPool.Get()
					if nil == natokHandler {
						logrus.Warnf("%s, Not find natokHandler!", sprintf)
						return nil
					}
					handler := &ConnectHandler{Conn: &DualConn{Packet: &Packet{PacketConn: packet, Addr: addr}}, ConnHandler: natokHandler}
					natokHandler.ConnHandler = handler
					return &ExtraConnectHandler{
						AccessId:    Udp + Protocol + addr.String(),
						AccessKey:   mapping.AccessKey,
						Sign:        mapping.PortSign,
						Port:        mapping.PortNum,
						ChanActive:  make(chan bool, 1),
						ConnHandler: handler,
					}
				})
				if extra == nil {
					return
				}
				extra.Activate()
				extra.ConnHandler.PacketRead(extra, buf, n)
			} else {
				logrus.Warnf("Not find accessKey: %s", mapping.AccessKey)
				_ = packet.Close()
			}
		}
	}(packet)
	return nil
}

func tcpListen(mapping *PortMapping) error {
	// 绑定服务端的端口
	listen, err := net.Listen(Tcp, util.ToAddress(mapping.PortScope, mapping.PortNum))
	if err != nil {
		logrus.Errorf("Listen tcp port %d failed! %v", mapping.PortNum, err.Error())
		return err
	}
	logrus.Infof("Bind tcp listen %s", listen.Addr())
	mapping.Listener = &DualListener{Listener: listen}
	go func(listener net.Listener) {
		// 服务端收到请求
		for {
			if mapping.Enabled == false {
				logrus.Infof("Port %d not enabled, exit tcp listen %s", mapping.PortNum, listener.Addr())
				_ = listener.Close()
				break
			}
			accept, err := listener.Accept()
			if err != nil {
				logrus.Errorf("Accept failed! %v", err.Error())
				break
			}
			// 限制开放名单
			if util.NoneMatch(mapping.Whitelist, func(ip string) bool {
				return strings.Contains(accept.RemoteAddr().String(), ip)
			}, false) {
				_ = accept.Close()
				continue
			}
			if cm, ifCM := ClientManage.Load(mapping.AccessKey); cm != nil && ifCM {
				// 客户端: 启用 && 在线
				client := cm.(*ClientBlocking)
				if !client.Enabled || client.NatokHandler == nil || !client.NatokHandler.Active {
					_ = accept.Close()
					continue
				}
				go func(accept net.Conn) {
					sprintf := fmt.Sprintf("Accept connect tcp://%s from %s %d", accept.RemoteAddr(), mapping.Protocol, mapping.PortNum)
					logrus.Debugf(sprintf)
					// 将外部连接与内部连接关联
					extra := getExtra(mapping.AccessKey, Tcp+Protocol+accept.RemoteAddr().String(), func() *ExtraConnectHandler {
						// 获取内部连接通道
						natokHandler := client.NatokPool.Get()
						if natokHandler == nil {
							logrus.Warnf("%s, Not find natokHandler!", sprintf)
							_ = accept.Close()
							return nil
						}
						handler := &ConnectHandler{Conn: &DualConn{Conn: accept}, ConnHandler: natokHandler}
						natokHandler.ConnHandler = handler
						return &ExtraConnectHandler{
							ConnHandler: handler,
							Sign:        mapping.PortSign,
							Port:        mapping.PortNum,
							ChanActive:  make(chan bool, 1),
							AccessKey:   mapping.AccessKey,
							AccessId:    Tcp + Protocol + accept.RemoteAddr().String(),
						}
					})
					if extra == nil {
						return
					}
					extra.Activate()
					extra.ConnHandler.Listen(extra)
					_ = accept.Close()
				}(accept)
			} else {
				logrus.Warnf("Not find accessKey: %s", mapping.AccessKey)
				_ = accept.Close()
			}
		}
	}(listen)
	return nil
}

// BindPort 绑定端口监听
func BindPort(mapping *PortMapping) error {
	if cm, ifCM := ClientManage.Load(mapping.AccessKey); cm != nil && ifCM {
		mapping.Enabled = true
		client := cm.(*ClientBlocking)
		if mapping.Protocol == Udp {
			if err := udpListen(mapping); err != nil {
				return err
			}
		} else {
			if err := tcpListen(mapping); err != nil {
				return err
			}
		}
		client.PortListener.Store(mapping.PortSign, mapping)
	}
	return nil
}

// UnBindPort 解除端口绑定
func UnBindPort(mapping *PortMapping) error {
	accessKey := mapping.AccessKey
	portSign := mapping.PortSign
	if cm, ifCM := ClientManage.Load(accessKey); cm != nil && ifCM {
		// 端口解绑
		client := cm.(*ClientBlocking)
		if pm, ifPM := client.PortListener.Load(portSign); pm != nil && ifPM {
			portMapping := pm.(*PortMapping)
			if nil != portMapping {
				portMapping.Enabled = false
				if portMapping.Listener != nil {
					// TCP
					if listen := portMapping.Listener.Listener; listen != nil {
						if err := listen.Close(); nil != err {
							logrus.Errorf("Unbind tcp port %s, %v", listen.Addr(), err)
							return err
						}
						logrus.Infof("Unbind tcp listen %s", listen.Addr())
					}
					// UDP
					if packet := portMapping.Listener.PacketConn; packet != nil {
						if err := packet.Close(); nil != err {
							logrus.Errorf("Unbind udp port %s, %v", packet.LocalAddr(), err)
							return err
						}
						logrus.Infof("Unbind udp listen %s", packet.LocalAddr())
					}
				}
			}
		}
		client.PortListener.Delete(portSign)
	}
	// 连接清除
	if cn, ifCN := ConnectManage.Load(accessKey); cn != nil && ifCN {
		blocking := cn.(*ConnectBlocking)
		if signs, ifSign := blocking.PortSignMap.Load(portSign); signs != nil && ifSign {
			for _, accessId := range signs.([]string) {
				if connect, ifConnect := blocking.AccessIdMap.Load(accessId); connect != nil && ifConnect {
					extra := connect.(*ExtraConnectHandler)
					if nil != extra && nil != extra.ConnHandler {
						// TCP
						if conn := extra.ConnHandler.Conn.Conn; conn != nil {
							_ = conn.Close()
						}
						// UDP
						if packet := extra.ConnHandler.Conn.Packet; packet != nil {
							packet.Addr = nil
							_ = packet.PacketConn.Close()
						}
					}
				}
				blocking.AccessIdMap.Delete(accessId)
			}
		}
		blocking.PortSignMap.Delete(portSign)
	}
	return nil
}
