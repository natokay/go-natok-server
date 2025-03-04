package util

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
	"strings"
	"time"
)

// PortAvailable 端口是否可用
func PortAvailable(port int, protocol string) bool {
	if strings.EqualFold("udp", protocol) {
		packet, err := net.ListenPacket("udp", ":"+strconv.Itoa(port))
		if err != nil {
			return false
		}
		defer func(packet net.PacketConn) {
			_ = packet.Close()
		}(packet)
	} else {
		listen, err := net.Listen("tcp", ":"+strconv.Itoa(port))
		if err != nil {
			return false
		}
		defer func(listen net.Listener) {
			_ = listen.Close()
		}(listen)
	}
	return true
}

// ToAddress scope=local时，返回本地，默认为全局
func ToAddress(scope string, port int) string {
	ip := "" // 默认为开放
	if scope == "local" {
		ip = "127.0.0.1"
	}
	return ip + ":" + strconv.Itoa(port)
}

// Md5 字符串md5加密处理
func Md5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// UUID V4 基于随机数
func UUID() string {
	return uuid.New().String()
}

// loc 时区Asia/Shanghai
var loc, _ = time.LoadLocation("Local")

// PreTs 前缀添加时间戳(纳秒)
func PreTs(s string) string {
	return time.Now().Format("20060102150405.000000000") + "." + s
}

// NowDayStart 当天的开始时间
func NowDayStart() time.Time {
	nowtDayStart, _ := time.ParseInLocation("2006-01-02", time.Now().Format("2006-01-02"), loc)
	return nowtDayStart
}

// DayStart 将目标时间转为当天的开始时间
func DayStart(target time.Time) time.Time {
	parse, err := time.ParseInLocation("2006-01-02", target.Format("2006-01-02"), loc)
	if err != nil {
		logrus.Errorf("%v", err.Error())
	}
	return parse
}

// Contains 检查特定字符串是否在切片中
func Contains(slice []string, element string) bool {
	for _, value := range slice {
		if value == element {
			return true
		}
	}
	return false
}

// RemoveIf 从切片中移除特定字符串
func RemoveIf(slice []string, element string) []string {
	var result []string
	for _, s := range slice {
		if s != element {
			result = append(result, s)
		}
	}
	return result
}

func NoneMatch(slice []string, predicate func(string) bool, def bool) bool {
	if slice != nil && len(slice) != 0 {
		for _, v := range slice {
			if predicate(v) {
				return false
			}
		}
		return true
	}
	return def
}

// Distinct 去重
func Distinct(s []string) []string {
	keys := make(map[string]bool)
	list := make([]string, 0)
	for _, entry := range s {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// GeneratePassword 生成指定长度的密码
func GeneratePassword(passLen int) (string, error) {
	// 检查密码长度是否小于6
	if passLen < 6 {
		return "", errors.New("password length must be not be less than 6")
	}
	// 定义包含3类字符的字符集（62个字符）
	charsetBytes := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	charsetLen := byte(len(charsetBytes))
	// 计算最大有效值以确保均匀分布
	maxValue := byte(255 - (256 % uint(charsetLen)))
	password := make([]byte, passLen)
	for i := 0; i < passLen; {
		buf := make([]byte, 1)
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}
		// 过滤超出范围的字节值
		if buf[0] <= maxValue {
			password[i] = charsetBytes[buf[0]%charsetLen]
			i++
		}
	}
	return string(password), nil
}
