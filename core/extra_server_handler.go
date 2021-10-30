package core

import (
	"github.com/kataras/golog"
	"go-natok-server/util"
	"net"
	"strconv"
	"time"
)

// ExtraServerHandler struct 外部端口请求服务处理
type ExtraServerHandler struct {
	AccessId    string //接受连接的ID
	AccessKey   string //绑定的访问秘钥
	Port        int    //绑定的端口
	Chan        chan *ConnectHandler
	ConnHandler *ConnectHandler
}

// Error 错误处理
func (e *ExtraServerHandler) Error(handler *ConnectHandler) {
	if load, ok := WorkServerGroupMap.Load(e.AccessKey); load != nil && ok {
		if nil != load.(*WorkBlocking).Heads[e.AccessId] {
			delete(load.(*WorkBlocking).Heads, e.AccessId)
		}
	}
	natokHandler := handler.ConnHandler
	if natokHandler != nil {
		natokHandler.Write(Message{Type: TypeDisconnect, Head: e.AccessId})
		natokHandler.Conn.Close()
	}
	handler.Conn.Close()
	handler.Active = false
}

// Encode 编码消息
func (e *ExtraServerHandler) Encode(msg interface{}) []byte {
	if msg == nil {
		return []byte{}
	}
	return msg.([]byte)
}

// Decode 解码消息
func (e *ExtraServerHandler) Decode(buf []byte) (interface{}, int) {
	return buf, len(buf)
}

// Receive 请求接收
func (e *ExtraServerHandler) Receive(handler *ConnectHandler, data interface{}) {
	// 判断客户端是否已连接，未连接则等待连接
	if handler.ConnHandler == nil {
		if load, ok := WorkServerGroupMap.Load(e.AccessKey); load != nil && ok {
			extraHandler := load.(*WorkBlocking).Heads[e.AccessId]
			select {
			case natokHandler := <-extraHandler.Chan:
				handler.ConnHandler = natokHandler
			case <-time.After(time.Second * 15):
				golog.Info("Port ", e.Port, " wait 15 second, connection timeout !")
				e.Error(handler)
				return
			}
		}
	}
	// 将请求转发到客户端
	if handler.ConnHandler != nil {
		msg := Message{Type: TypeTransfer, Body: data.([]byte)}
		handler.ConnHandler.Write(msg)
	}
}

// Active 激活（将内部网地址发送给客户端建立连接）
func (e *ExtraServerHandler) Active(handler *ConnectHandler) {
	e.Port = handler.Conn.LocalAddr().(*net.TCPAddr).Port
	// 若外部端口已完成绑定，激活该请求
	ClientGroupManage.Range(func(key, value interface{}) bool {
		client := value.(*ClientBlocking)
		for _, mapping := range client.Mapping {
			if e.Port == mapping.Port {
				err, snowflakeId := util.DefaultSnowflakeId()
				if nil != err {
					golog.Error("雪花算法生成ID错误", err)
					return false
				}
				e.AccessId = strconv.FormatInt(snowflakeId, 10)
				e.AccessKey = mapping.AccessKey

				if nil != client.Handler {
					if work, ok := WorkServerGroupMap.Load(e.AccessKey); work != nil && ok {
						work.(*WorkBlocking).Heads[e.AccessId] = e
						work.(*WorkBlocking).Signs[mapping.Sign] = append(work.(*WorkBlocking).Signs[mapping.Sign], e.AccessId)
					} else {
						WorkServerGroupMap.Store(e.AccessKey, &WorkBlocking{
							Signs: map[string][]string{mapping.Sign: {e.AccessId}},
							Heads: map[string]*ExtraServerHandler{e.AccessId: e},
						})
					}
					msg := Message{Type: TypeConnect, Head: e.AccessId, Body: []byte(mapping.Intranet)}
					client.Handler.Write(msg)
					golog.Info("Active connect port ", e.Port, " access id->", e.AccessId, " access key->", e.AccessKey)
					return false
				}
			}
		}
		return true
	})
	// 校验外部端口是否已完成绑定
	if e.AccessKey == "" {
		handler.Conn.Close()
		handler.Conn = nil
	}
}

// BindPort 绑定端口监听
func BindPort(mapping PortMapping) error {
	if value, ok := ClientGroupManage.Load(mapping.AccessKey); value != nil && ok {
		if value.(*ClientBlocking).Enabled == true {

			listen, err := net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(mapping.Port))
			if err != nil {
				golog.Error("端口监听失败！port={}", mapping.Port, err.Error())
				return err
			}
			golog.Info("Bind listen", listen.Addr())

			go func(listener net.Listener) {
				for {
					accept, err := listener.Accept()
					if err != nil {
						golog.Error("Error: 端口监听建立连接失败！", err.Error())
						break
					}
					go func(conn net.Conn) {
						handler := &ConnectHandler{Conn: conn}
						msgHandler := &ExtraServerHandler{ConnHandler: handler, Chan: make(chan *ConnectHandler, 1)}
						msgHandler.Active(handler)
						handler.Listen(conn, msgHandler)
					}(accept)
				}
			}(listen)

			mapping.Listener = listen
			value.(*ClientBlocking).Mapping[mapping.Sign] = &mapping
		}
	}
	return nil
}

// UnBindPort 解除端口绑定
func UnBindPort(mapping PortMapping) error {
	accessKey := mapping.AccessKey
	sign := mapping.Sign
	if value, ok := ClientGroupManage.Load(accessKey); value != nil && ok {

		clientBlocking := value.(*ClientBlocking)
		portMapping := clientBlocking.Mapping[sign]

		if nil != portMapping && nil != portMapping.Listener {
			listen := portMapping.Listener
			if err := listen.Close(); nil != err {
				golog.Error("Error! Unbind Port", listen.Addr(), err)
				return err
			}
			golog.Info("Unbind listen", listen.Addr())

			if work, ok := WorkServerGroupMap.Load(accessKey); work != nil && ok {
				for _, sn := range work.(*WorkBlocking).Signs[sign] {
					extra := work.(*WorkBlocking).Heads[sn]
					if nil != extra && nil != extra.ConnHandler {
						extra.ConnHandler.Conn.Close()
						extra.ConnHandler.Active = false
					}
				}
			}
			WorkServerGroupMap.Delete(accessKey)
			delete(clientBlocking.Mapping, sign)
			return nil
		}
	}
	return nil
}
