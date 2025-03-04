package core

import (
	"github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

var (
	ClientManage   sync.Map //Natok客户端(AccessKey,*ClientBlocking)
	ConnectManage  sync.Map //外部请求服务(AccessKey,*ConnectBlocking)
	ChanClientSate = make(chan ChanClient, 100)
)

// ConnectBlocking struct 连接
type ConnectBlocking struct {
	PortSignMap sync.Map //映射签名<PortSign,[AccessId1,AccessId2]>
	AccessIdMap sync.Map //连接句柄<AccessId,*ExtraConnectHandler>
}

// ClientBlocking struct 客户端通道对象
type ClientBlocking struct {
	Enabled      bool            //是否可用
	AccessKey    string          //客户端秘钥
	NatokPool    *ConnectPool    //客户端连接池
	NatokHandler *ConnectHandler //客户端连接句柄
	PortListener sync.Map        //PortSign -> *PortMapping
}

// DualListener struct TCP、UDP监听器
type DualListener struct {
	net.Listener
	net.PacketConn
}

// ChanClient struct C-S连接状态
type ChanClient struct {
	AccessKey string
	State     int8
}

// PortMapping struct 端口映射对象
type PortMapping struct {
	Enabled     bool            //是否可用
	AccessKey   string          //访问秘钥
	PortSign    string          //映射签名
	PortScope   string          //监听范围
	PortNum     int             //访问端口
	Intranet    string          //转发地址
	Protocol    string          //协议类型
	Whitelist   []string        //开放名单
	Listener    *DualListener   //ServerListener
	ConnHandler *ConnectHandler //连接句柄
}

// ConnectHandler struct 通道链接载体
type ConnectHandler struct {
	ReadTime    time.Time       //读取时间
	WriteTime   time.Time       //写入时间
	Active      bool            //是否活跃
	primary     bool            //核心通道
	ReadBuf     []byte          //读取的内容
	Conn        *DualConn       //连接通道
	MsgHandler  MsgHandler      //消息句柄
	ConnHandler *ConnectHandler //连接句柄
}

// DualConn struct TCP、UDP连接器
type DualConn struct {
	net.Conn
	*Packet
}

// Packet UDP
type Packet struct {
	net.PacketConn
	net.Addr
}

// Message 内部通信消息体
type Message struct {
	Type   byte   //消息类型
	Serial string //消息序列
	Net    string //网络类型
	Uri    string //消息头
	Data   []byte //消息体
}

// MsgHandler interface 消息处理接口
type MsgHandler interface {
	Encode(interface{}) []byte            //加密
	Decode([]byte) (interface{}, int)     //解密
	Receive(*ConnectHandler, interface{}) //接收
	Close(*ConnectHandler)                //关闭
}

// Write ConnectHandler 消息写入
func (c *ConnectHandler) Write(msg interface{}) {
	if c.MsgHandler == nil {
		return
	}
	data := c.MsgHandler.Encode(msg)
	c.WriteTime = time.Now()
	// TCP
	if conn := c.Conn.Conn; conn != nil {
		if _, err := conn.Write(data); err != nil {
			logrus.Errorf("%v", err.Error())
		}
	}
	// UDP
	if packet := c.Conn.Packet; packet != nil {
		if _, err := packet.WriteTo(data, packet.Addr); err != nil {
			logrus.Errorf("%v", err.Error())
		}
	}
}

// NatokClient 客户连接监听
func NatokClient(listener net.Listener) {
	for {
		accept, err := listener.Accept()
		if err != nil {
			logrus.Errorf("Natok client listen failed! %s, %v", listener.Addr(), err)
			continue
		}
		go func(accept net.Conn) {
			handler := &ConnectHandler{Conn: &DualConn{Conn: accept}}
			handler.Listen(&NatokConnectHandler{ConnHandler: handler})
			_ = accept.Close()
		}(accept)
	}
}

// Listen 连接请求监听
func (c *ConnectHandler) Listen(msgHandler interface{}) {
	defer func() {
		c.Active = false
		if err := recover(); err != nil {
			//debug.PrintStack()
			c.MsgHandler.Close(c)
			logrus.Warn(err)
		}
	}()
	c.Active = true
	c.ReadTime = time.Now()
	c.MsgHandler = msgHandler.(MsgHandler)

	for c.Active {
		// 接收数据 tcp buffer size 16kb
		buf := make([]byte, 1024*16)
		if c.ReadBuf != nil && len(c.ReadBuf) > MaxPacketSize {
			logrus.Warn("This conn is error ! Packet max than 4M !")
			c.MsgHandler.Close(c)
			break
		}
		// 检查连接是否为 nil
		if c.Conn == nil || c.Conn.Conn == nil {
			logrus.Error("This conn is nil")
			c.MsgHandler.Close(c)
			break
		}
		// 从连接读取数据
		n, err := c.Conn.Conn.Read(buf)
		if err != nil || n == 0 {
			if err != nil && err.Error() != "EOF" {
				logrus.Errorf("%v", err.Error())
			}
			c.MsgHandler.Close(c)
			break
		}

		c.ReadTime = time.Now()
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
		// 在包未读取完成时，需要二次读取
		if len(c.ReadBuf) > 0 {
			buf := make([]byte, len(c.ReadBuf))
			copy(buf, c.ReadBuf)
			c.ReadBuf = buf
		}
	}
}

// PacketRead 连接请求监听
func (c *ConnectHandler) PacketRead(msgHandler MsgHandler, buf []byte, n int) {
	defer func() {
		c.Active = false
		if err := recover(); err != nil {
			//debug.PrintStack()
			c.MsgHandler.Close(c)
			logrus.Warn(err)
		}
	}()
	c.Active = true
	c.ReadTime = time.Now()
	c.MsgHandler = msgHandler
	c.ReadTime = time.Now()
	c.ReadBuf = buf[0:n]

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
}
