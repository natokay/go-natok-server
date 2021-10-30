package util

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/google/uuid"
	"github.com/kataras/golog"
	"time"
)

//字符串md5加密处理
func Md5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

//V4 基于随机数
func UUID() string {
	return uuid.New().String()
}

// loc 时区Asia/Shanghai
var loc, _ = time.LoadLocation("Local")

// NowDayStart 当天的开始时间
func NowDayStart() time.Time {
	nowtDayStart, _ := time.ParseInLocation("2006-01-02", time.Now().Format("2006-01-02"), loc)
	return nowtDayStart
}

// DayStart 将目标时间转为当天的开始时间
func DayStart(target time.Time) time.Time {
	parse, err := time.ParseInLocation("2006-01-02", target.Format("2006-01-02"), loc)
	if err != nil {
		golog.Error(err)
	}
	return parse
}
