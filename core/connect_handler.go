package core

import (
	"github.com/kataras/golog"
	"net"
	"sync"
	"time"
)

var (
	ClientGroupManage  sync.Map //Natok客户端(AccessKey,*ClientBlocking)
	WorkServerGroupMap sync.Map //外部请求服务(AccessKey,*WorkBlocking)
	ChanClientSate     = make(chan ChanClient, 100)
)

// MsgHandler interface 消息处理接口
type MsgHandler interface {
	Error(*ConnectHandler)                //出错
	Encode(interface{}) []byte            //加密
	Decode([]byte) (interface{}, int)     //解密
	Receive(*ConnectHandler, interface{}) //接收
}

// Message struct 内部通信消息体
type Message struct {
	Type      byte
	SerialNum uint64
	Head      string
	Body      []byte
}

// ConnectHandler struct 通道链接载体
type ConnectHandler struct {
	ReadTime    int64           //读取时间
	WriteTime   int64           //写入时间
	Active      bool            //是否活跃
	primary     bool            //核心通道
	ReadBuf     []byte          //读取的内容
	Conn        net.Conn        //连接通道
	MsgHandler  MsgHandler      //消息句柄
	ConnHandler *ConnectHandler //连接句柄
}

// ClientBlocking struct 客户端通道对象
type ClientBlocking struct {
	AccessKey string                  //访问秘钥
	Enabled   bool                    //是否可用
	Handler   *ConnectHandler         //连接句柄
	Listener  net.Listener            //ClientListener
	Mapping   map[string]*PortMapping //Sign -> PortMapping
}

// PortMapping struct 端口映射对象
type PortMapping struct {
	AccessKey string       //访问秘钥
	Sign      string       //映射签名
	Port      int          //暴露端口
	Intranet  string       //内部地址
	Domain    string       //绑定域名
	Protocol  string       //协议类型
	Listener  net.Listener //ServerListener
}

// WorkBlocking struct 连接
type WorkBlocking struct {
	Signs map[string][]string            //映射签名<sign,[AccessId1,AccessId2]>
	Heads map[string]*ExtraServerHandler //连接句柄<AccessId,*ExtraServerHandler>
}

// ChanClient struct C-S连接状态
type ChanClient struct {
	AccessKey string
	State     int
}

// Write ConnectHandler 消息写入
func (c *ConnectHandler) Write(msg interface{}) {
	if c.MsgHandler == nil {
		return
	}
	data := c.MsgHandler.Encode(msg)
	c.WriteTime = time.Now().Unix()
	c.Conn.Write(data)
}

// Listen 连接请求监听
func (c *ConnectHandler) Listen(conn net.Conn, msgHandler interface{}) {
	defer func() {
		if err := recover(); err != nil {
			golog.Warn("Warn: %v\n", err)
			//debug.PrintStack()
			//c.MsgHandler.Error(c)
		}
	}()

	if conn == nil {
		return
	}

	c.Conn = conn
	c.Active = true
	c.ReadTime = time.Now().Unix()
	c.MsgHandler = msgHandler.(MsgHandler)

	for c.Active {
		buf := make([]byte, 1024*64)
		if c.ReadBuf != nil && len(c.ReadBuf) > MaxPacketSize {
			golog.Warn("Warn:  This conn is error ! Packet max than 4M !")
			c.MsgHandler.Error(c)
			break
		}

		n, err := c.Conn.Read(buf)
		if err != nil || n == 0 {
			golog.Error("Error:%v", err)
			//debug.PrintStack()
			c.MsgHandler.Error(c)
			break
		}

		c.ReadTime = time.Now().Unix()
		if c.ReadBuf == nil {
			c.ReadBuf = buf[0:n]
		} else {
			c.ReadBuf = append(c.ReadBuf, buf[0:n]...)
		}

		for {
			msg, n := c.MsgHandler.Decode(c.ReadBuf)
			if msg == nil {
				break
			}
			c.MsgHandler.Receive(c, msg)
			c.ReadBuf = c.ReadBuf[n:]
			if len(c.ReadBuf) == 0 {
				break
			}
		}

		if len(c.ReadBuf) > 0 {
			buf := make([]byte, len(c.ReadBuf))
			copy(buf, c.ReadBuf)
			c.ReadBuf = buf
		}
	}
}
