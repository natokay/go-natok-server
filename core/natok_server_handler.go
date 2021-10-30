package core

import (
	"encoding/binary"
	"github.com/kataras/golog"
	"strings"
	"time"
)

// NatokServerHandler struct NATOK服务端请求服务处理
type NatokServerHandler struct {
	accessId    string
	accessKey   string
	ConnHandler *ConnectHandler
}

// Error 错误处理
func (n *NatokServerHandler) Error(handler *ConnectHandler) {
	if handler.primary {
		ChanClientSate <- ChanClient{AccessKey: n.accessKey, State: 0}
	}
	n.handleDisconnect(handler)
}

// Encode 编码消息
func (n *NatokServerHandler) Encode(inMsg interface{}) []byte {
	if inMsg == nil {
		return []byte{}
	}
	msg := inMsg.(Message)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, msg.SerialNum)

	headBytes := []byte(msg.Head)
	bodyLen := TypeSize + SerialNumSize + HeadLenSize + len(headBytes) + len(msg.Body)

	data := make([]byte, HeaderSize, bodyLen+HeaderSize)
	binary.BigEndian.PutUint32(data, uint32(bodyLen))

	data = append(data, msg.Type)
	data = append(data, buf...)
	data = append(data, byte(len(headBytes)))
	data = append(data, headBytes...)
	data = append(data, msg.Body...)
	return data
}

// Decode 解码消息
func (n *NatokServerHandler) Decode(buf []byte) (interface{}, int) {
	headerBytes := buf[0:HeaderSize]
	headerLen := binary.BigEndian.Uint32(headerBytes)

	if uint32(len(buf)) < headerLen+HeaderSize {
		return nil, 0
	}

	num := int(headerLen + HeaderSize)
	body := buf[HeaderSize:num]

	headLen := uint8(body[SerialNumSize+TypeSize])
	msg := Message{
		Type:      body[0],
		SerialNum: binary.BigEndian.Uint64(body[TypeSize : SerialNumSize+TypeSize]),
		Head:      string(body[SerialNumSize+TypeSize+HeadLenSize : SerialNumSize+TypeSize+HeadLenSize+headLen]),
		Body:      body[SerialNumSize+TypeSize+HeadLenSize+headLen:],
	}
	return msg, num
}

// Receive 请求接收
func (n *NatokServerHandler) Receive(handler *ConnectHandler, msgData interface{}) {
	msg := msgData.(Message)
	switch msg.Type {
	case TypeAuth:
		n.handleAuthor(handler, msg)
	case TypeConnect:
		n.handleConnect(handler, msg)
	case TypeTransfer:
		n.handleTransfer(handler, msg)
	case TypeHeartbeat:
		n.handleHeartbeat(handler, msg)
	case TypeDisconnect:
		n.handleDisconnect(handler)
	}
}

// handleDisconnect 断开连接
func (n *NatokServerHandler) handleDisconnect(handler *ConnectHandler) {
	if load, ok := WorkServerGroupMap.Load(n.accessKey); load != nil && ok {
		signs := load.(*WorkBlocking).Signs
		for key, accessIds := range signs {
			for i, id := range accessIds {
				if n.accessId == id {
					accessIds = append(accessIds[:i], accessIds[i+1:]...)
				}
			}
			signs[key] = accessIds
		}
		delete(load.(*WorkBlocking).Heads, n.accessId)
	}
	extraHandler := handler.ConnHandler
	if extraHandler != nil && extraHandler.Conn != nil {
		extraHandler.Conn.Close()
	}
	handler.Conn.Close()
	handler.Active = false

	golog.Info("Disconnect access id->", n.accessId, " access key->", n.accessKey)
}

// handleConnect 建立连接
func (n *NatokServerHandler) handleConnect(handler *ConnectHandler, msg Message) {
	head := msg.Head
	if head == "" {
		golog.Warn("Warn the head is empty.")
		n.Error(handler)
		return
	}
	tokens := strings.Split(head, "@")
	if len(tokens) < 2 {
		golog.Warn("Warn the head is not valid.")
		n.Error(handler)
		return
	}
	n.accessId = tokens[0]
	n.accessKey = tokens[1]
	//寻找外部请求，将其与来自客户端的请求绑定在一起
	if work, ok := WorkServerGroupMap.Load(n.accessKey); work != nil && ok {
		extraHandler := work.(*WorkBlocking).Heads[n.accessId]
		extraHandler.Chan <- handler
		handler.ConnHandler = extraHandler.ConnHandler
	}

	golog.Info("Accept connect access id->", n.accessId, "access key->", n.accessKey)
}

// handleTransfer 数据传输
func (n *NatokServerHandler) handleTransfer(handler *ConnectHandler, msg Message) {
	handler.ConnHandler.Write(msg.Body)
}

// handleHeartbeat Natok客户端心跳
func (n *NatokServerHandler) handleHeartbeat(handler *ConnectHandler, msg Message) {
	handler.Write(msg)
}

// handleAuthor Natok客户端认证
func (n *NatokServerHandler) handleAuthor(handler *ConnectHandler, msg Message) {
	n.accessKey = msg.Head
	value, ok := ClientGroupManage.Load(n.accessKey)
	if !ok {
		msg.Type = TypeInvalidKey
		handler.Write(msg)
		n.Error(handler)
		return
	}
	if nil == value {
		msg.Type = TypeDisabledAccessKey
		handler.Write(msg)
		n.Error(handler)
		return
	}
	//判断是否已启用
	clientBlocking := value.(*ClientBlocking)
	if !clientBlocking.Enabled {
		msg.Type = TypeDisabledAccessKey
		handler.Write(msg)
		n.Error(handler)
		return
	}
	//判断是否有绑定端口
	if len(clientBlocking.Mapping) <= 0 {
		msg.Type = typeNoAvailablePort
		handler.Write(msg)
		golog.Warnf("There are no available ports for the natok access key [%s]", n.accessKey)
	}
	//判断是否前面已建立过连接
	if clientBlocking.Handler != nil && clientBlocking.Handler.Conn != handler.Conn {
		clientBlocking.Handler.primary = false
		msg.Type = TypeIsInuseKey
		clientBlocking.Handler.Conn.Write(n.Encode(msg))
		golog.Warn("The access key [", n.accessKey, "] is in use by other natok client", clientBlocking.Handler.Conn.RemoteAddr(), "->", handler.Conn.RemoteAddr())
	}

	handler.primary = true
	clientBlocking.Handler = handler
	ClientGroupManage.Store(n.accessKey, clientBlocking)

	ChanClientSate <- ChanClient{AccessKey: n.accessKey, State: 1}
	golog.Info("The access key [", n.accessKey, "] with ports", len(clientBlocking.Mapping), "is in use by natok client", handler.Conn.RemoteAddr(), " online at", time.Now().Format("2006-01-02 15:04:05"))
}
