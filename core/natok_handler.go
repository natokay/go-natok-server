package core

import (
	"encoding/binary"
	"github.com/sirupsen/logrus"
	"natok-server/support"
	"time"
)

// NatokConnectHandler struct NATOK服务端请求服务处理
type NatokConnectHandler struct {
	Serial      string //序列
	AccessId    string //接受连接的ID
	AccessKey   string //绑定的客户端秘钥
	Sign        string //映射签名
	ConnHandler *ConnectHandler
}

// Close 关闭处理
func (s *NatokConnectHandler) Close(handler *ConnectHandler) {
	if handler.primary {
		// 客户端离线
		ChanClientSate <- ChanClient{AccessKey: s.AccessKey, State: 0}
		// 关闭客户端连接池
		if cm, ifCM := ClientManage.Load(s.AccessKey); cm != nil && ifCM {
			client := cm.(*ClientBlocking)
			client.NatokPool.Shutdown()
		}
	}
	handler.Active = false
	s.disconnect(handler)
}

// Encode 编码消息
func (s *NatokConnectHandler) Encode(inMsg interface{}) []byte {
	if inMsg == nil {
		return []byte{}
	}
	msg := inMsg.(Message)
	serialBytes := []byte(msg.Serial)
	netBytes := []byte(msg.Net)
	UriBytes := []byte(msg.Uri)
	// byte=Uint8Size,3个string=Uint8Size*3,+data
	dataLen := Uint8Size + Uint8Size*3 + len(serialBytes) + len(netBytes) + len(UriBytes) + len(msg.Data)
	data := make([]byte, Uint32Size, Uint32Size+dataLen)
	binary.BigEndian.PutUint32(data, uint32(dataLen))

	data = append(data, msg.Type)
	data = append(data, byte(len(serialBytes)))
	data = append(data, byte(len(netBytes)))
	data = append(data, byte(len(UriBytes)))
	data = append(data, serialBytes...)
	data = append(data, netBytes...)
	data = append(data, UriBytes...)
	data = append(data, msg.Data...)
	return data
}

// Decode 解码消息
func (s *NatokConnectHandler) Decode(buf []byte) (interface{}, int) {
	headerBytes := buf[0:Uint32Size]
	headerLen := binary.BigEndian.Uint32(headerBytes)
	// 来自客户端的包，校验完整性。
	if uint32(len(buf)) < headerLen+Uint32Size {
		return nil, 0
	}

	head := int(Uint32Size + headerLen)
	body := buf[Uint32Size:head]
	serialLen := int(body[Uint8Size])
	netLen := int(body[Uint8Size*2])
	uriLen := int(body[Uint8Size*3])
	msg := Message{
		Type:   body[0],
		Serial: string(body[Uint8Size*4 : Uint8Size*4+serialLen]),
		Net:    string(body[Uint8Size*4+serialLen : Uint8Size*4+serialLen+netLen]),
		Uri:    string(body[Uint8Size*4+serialLen+netLen : Uint8Size*4+serialLen+netLen+uriLen]),
		Data:   body[Uint8Size*4+serialLen+netLen+uriLen:],
	}
	return msg, head
}

// Receive 请求接收
func (s *NatokConnectHandler) Receive(handler *ConnectHandler, msgData interface{}) {
	msg := msgData.(Message)
	switch msg.Type {
	case TypeAuth:
		s.author(handler, msg)
	case TypeConnectNatok:
		s.connect(handler, msg)
	case TypeConnectIntra:
		s.intra(handler, msg)
	case TypeTransfer:
		s.transfer(handler, msg)
	case TypeHeartbeat:
		s.heartbeat(handler, msg)
	case TypeDisconnect:
		s.disconnect(handler)
	}
}

// author Natok客户端认证
func (s *NatokConnectHandler) author(handler *ConnectHandler, msg Message) {
	s.AccessKey = msg.Uri
	cm, ifCM := ClientManage.Load(s.AccessKey)
	// 无效的访问密钥
	if cm == nil || !ifCM {
		msg.Type = TypeInvalidKey
		handler.Write(msg)
		s.Close(handler)
		return
	}
	// 客户端启用检查
	client := cm.(*ClientBlocking)
	if !client.Enabled {
		msg.Type = TypeDisabledAccessKey
		handler.Write(msg)
		logrus.Warnf("The AccessKey [%s] is disabled", s.AccessKey)
	}
	// 绑定端口检查
	if IsEmpty(&client.PortListener) {
		msg.Type = typeNoAvailablePort
		handler.Write(msg)
		logrus.Warnf("The AccessKey [%s] no port available", s.AccessKey)
	}

	// 判断是否前面已建立过连接
	if client.NatokHandler != nil && client.NatokHandler != handler {
		client.NatokHandler.primary = false
		msg.Type = TypeIsInuseKey
		_, _ = client.NatokHandler.Conn.Write(s.Encode(msg))
		logrus.Warnf("The accessKey [%s] use by other natok client %s -> %s", s.AccessKey, client.NatokHandler.Conn.RemoteAddr(), handler.Conn.RemoteAddr())
	}
	// 标记为主连接
	handler.primary = true
	client.NatokHandler = handler
	ClientManage.Store(s.AccessKey, client)
	// 建立连接池
	pool := support.AppConf.Natok.Server.ChanPool
	client.NatokPool = NewConnectionPool(pool.MinSize, pool.MaxSize, time.Duration(pool.IdleTimeout)*time.Second, handler)
	// 客户端上线
	ChanClientSate <- ChanClient{AccessKey: s.AccessKey, State: 1}
	logrus.Infof("The accessKey [%s] with ports %d in natok client %s online at %s", s.AccessKey, GetLen(&client.PortListener), handler.Conn.RemoteAddr(), time.Now().Format("2006-01-02 15:04:05"))
}

// connect 建立连接
func (s *NatokConnectHandler) connect(handler *ConnectHandler, msg Message) {
	s.Serial = msg.Serial
	s.AccessKey = msg.Uri
	if s.AccessKey == "" {
		logrus.Warn("The AccessKey is empty")
		s.Close(handler)
		return
	}
	// 放入连接池
	if handler.Active {
		if cm, ifCM := ClientManage.Load(s.AccessKey); cm != nil && ifCM {
			client := cm.(*ClientBlocking)
			client.NatokPool.Accept(handler)
			return
		}
	}
}

// intra 建立内部连接
func (s *NatokConnectHandler) intra(handler *ConnectHandler, msg Message) {
	if extra := handler.ConnHandler; extra != nil {
		extra.MsgHandler.(*ExtraConnectHandler).ChanActive <- true
	}
}

// transfer 数据传输
func (s *NatokConnectHandler) transfer(handler *ConnectHandler, msg Message) {
	handler.ConnHandler.Write(msg.Data)
}

// heartbeat Natok客户端心跳
func (s *NatokConnectHandler) heartbeat(handler *ConnectHandler, msg Message) {
	// 心跳不改变写入时间
	wt := handler.WriteTime
	handler.Write(msg)
	handler.WriteTime = wt
}

// disconnect 断开连接
func (s *NatokConnectHandler) disconnect(handler *ConnectHandler) {
	// 关闭外部端口连接
	extraConnHandler := handler.ConnHandler
	if extraConnHandler != nil && extraConnHandler.Conn != nil {
		// TCP
		if conn := extraConnHandler.Conn.Conn; conn != nil {
			_ = conn.Close()
		}
		// UDP
		if packet := extraConnHandler.Conn.Packet; packet != nil {
			packet.Addr = nil
		}
		handler.ConnHandler = nil
		logrus.Debugf("Disconnect accessKey %s -> %s ", s.AccessKey, s.Serial)
	}
	// 放入连接池
	if handler.Active {
		if cm, ifCM := ClientManage.Load(s.AccessKey); cm != nil && ifCM {
			client := cm.(*ClientBlocking)
			client.NatokPool.Put(handler)
			return
		}
	}
}
