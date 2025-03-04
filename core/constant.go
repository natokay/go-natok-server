package core

import "sync"

// 关键字常量
const (
	Protocol = "://"
	Sqlite   = "sqlite"
	Mysql    = "mysql"
	Empty    = ""
)

// NATOK网络转发类型
const (
	Tcp      = "tcp"
	Udp      = "udp"
	Http     = "http"
	Https    = "https"
	Ssh      = "ssh"
	Ftp      = "ftp"
	Database = "data base"
	Desktop  = "remote desktop"
)

// 数据包常量
const (
	Uint8Size     = 1
	Uint16Size    = 2
	Uint32Size    = 4
	Uint64Size    = 8
	MaxPacketSize = 4 * 1 << 20 // 最大数据包大小为最大数据包大小为 4M
)

// 消息类型常量
const (
	TypeAuth                = 0x01 //验证消息以检查访问密钥是否正确
	typeNoAvailablePort     = 0x02 //访问密钥没有可用端口
	TypeConnectNatok        = 0xa1 //连接到NATOK服务
	TypeConnectIntra        = 0xa2 //连接到内部服务
	TypeDisconnect          = 0x04 //断开
	TypeTransfer            = 0x05 //数据传输
	TypeIsInuseKey          = 0x06 //客户端秘钥已在其他客户端使用
	TypeHeartbeat           = 0x07 //心跳
	TypeDisabledAccessKey   = 0x08 //禁用的访问密钥
	TypeDisabledTrialClient = 0x09 //禁用的试用客户端
	TypeInvalidKey          = 0x10 //无效的访问密钥
)

// IsEmpty 是否为空
func IsEmpty(m *sync.Map) bool {
	ifBool := true
	m.Range(func(_, _ interface{}) bool {
		ifBool = false
		return false
	})
	return ifBool
}

// IsNotEmpty 是否不为空
func IsNotEmpty(m *sync.Map) bool {
	return !IsEmpty(m)
}

// GetLen 获取长度
func GetLen(m *sync.Map) int {
	counter := 0
	m.Range(func(_, _ interface{}) bool {
		counter++
		return true
	})
	return counter
}
