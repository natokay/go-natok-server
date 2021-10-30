package core

// NATOK网络转发类型
const (
	Http     = "http"
	Https    = "https"
	Tcp      = "tcp"
	Ssh      = "ssh"
	Telnet   = "telnet"
	Database = "data base"
	Desktop  = "remote desktop"
)

// 数据包常量
const (
	TypeSize      = 1
	HeaderSize    = 4
	HeadLenSize   = 1
	SerialNumSize = 8
	MaxPacketSize = 4 * 1 << 20 // 最大数据包大小为最大数据包大小为 4M
)

// 消息类型常量
const (
	TypeAuth                = 0x01 // 验证消息以检查访问密钥是否正确
	typeNoAvailablePort     = 0x02 // 访问密钥没有可用端口
	TypeConnect             = 0x03 //  连接
	TypeDisconnect          = 0x04 //  断开
	TypeTransfer            = 0x05 //  数据传输
	TypeIsInuseKey          = 0x06 // 访问秘钥已在其他客户端使用
	TypeHeartbeat           = 0x07 // 心跳
	TypeDisabledAccessKey   = 0x08 // 禁用的访问密钥
	TypeDisabledTrialClient = 0x09 // 禁用的试用客户端
	TypeInvalidKey          = 0x10 // 无效的访问密钥
)
