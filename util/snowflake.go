package util

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

var (
	// 开始计时时间
	epoch = func() int64 { return time.Now().UnixNano() / int64(time.Millisecond) }
	// 时间戳最大值
	// timestampMax       int64 = -1 ^ (-1 << 41)
	snowflakeWorker    *IdWorker
	workerBits               = uint(5)
	dataCenterBits           = uint(5)
	sequenceBits             = uint(12)
	maxWorkerId        int64 = -1 ^ (-1 << workerBits)
	maxDataCenterId    int64 = -1 ^ (-1 << dataCenterBits)
	workerIdShift            = sequenceBits
	dataCenterIdShift        = sequenceBits + workerBits
	timestampLeftShift       = sequenceBits + workerBits + dataCenterBits
	sequenceMask       int64 = -1 ^ (-1 << sequenceBits)
)

type IdWorker struct {
	mutex         sync.Mutex
	workerId      int64
	dataCenterId  int64
	epoch         int64
	sequence      int64
	lastTimestamp int64
}

// DefSnowflakeId 默认获取雪花ID
func DefSnowflakeId() (error, int64) {
	return SnowflakeId(0, 0)
}

// DefStrSnowflakeId 默认获取雪花ID
func DefStrSnowflakeId() (error, string) {
	return StrSnowflakeId(0, 0)
}

// SnowflakeId 获取雪花ID
func SnowflakeId(workerId, dataCenterId int64) (error, int64) {
	var err error
	if nil == snowflakeWorker {
		err, snowflakeWorker = NewIdWorker(workerId, dataCenterId, epoch())
	}
	if err == nil {
		return snowflakeWorker.NextId()
	}
	return err, 0
}

// StrSnowflakeId 获取雪花ID
func StrSnowflakeId(workerId, dataCenterId int64) (error, string) {
	if err, id := SnowflakeId(workerId, dataCenterId); err == nil {
		return nil, strconv.FormatInt(id, 10)
	} else {
		return err, ""
	}
}

// NewIdWorker new worker
func NewIdWorker(workerId, dataCenterId, epoch int64) (error, *IdWorker) {
	if workerId > maxWorkerId || workerId < 0 {
		return errors.New(fmt.Sprintf("worker Id can't be greater than %d or less than 0", maxWorkerId)), nil
	}
	if dataCenterId > maxDataCenterId || dataCenterId < 0 {
		return errors.New(fmt.Sprintf("datacenter Id can't be greater than %d or less than 0", maxDataCenterId)), nil
	}
	id := &IdWorker{}
	id.workerId = workerId
	id.dataCenterId = dataCenterId
	id.sequence = 0
	id.lastTimestamp = -1
	id.epoch = epoch
	id.mutex = sync.Mutex{}
	return nil, id
}

// NextId 生成一个ID
func (i *IdWorker) NextId() (error, int64) {
	// 加锁, 防止数据被更改
	i.mutex.Lock()
	defer i.mutex.Unlock()
	timestamp := i.GenTime()
	// 如果时间出现回拨, 直接抛弃
	if timestamp < i.lastTimestamp {
		// "clock is moving backwards. Rejecting requests until %d.", lastTimestamp
		err := errors.New(fmt.Sprintf("Clock moved backwards. Refusing to generate id for %d milliseconds", i.lastTimestamp-timestamp))
		return err, 0
	}
	// 同一毫秒内出现多个请求, sequence +1
	// 最大值是4095, 超过4095则调用tilNextMillis进行等待
	if i.lastTimestamp == timestamp {
		i.sequence = (i.sequence + 1) & sequenceMask
		if i.sequence == 0 {
			timestamp = i.tilNextMillis(i.lastTimestamp)
		}
	} else {
		i.sequence = 0
	}
	i.lastTimestamp = timestamp
	return nil, ((timestamp - i.epoch) << timestampLeftShift) |
		(i.dataCenterId << dataCenterIdShift) |
		(i.workerId << workerIdShift) |
		i.sequence
}

// tilNextMillis 等待下一毫秒
func (i *IdWorker) tilNextMillis(lastTimestamp int64) int64 {
	timestamp := i.GenTime()
	for timestamp <= lastTimestamp {
		timestamp = i.GenTime()
	}
	return timestamp
}

// GenTime 获取当前时间, 单位是毫秒
func (i *IdWorker) GenTime() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// GetWorkerId 获得workerID
func (i *IdWorker) GetWorkerId() int64 {
	return i.workerId
}

// GetDataCenterID 获得datacenterID
func (i *IdWorker) GetDataCenterID() int64 {
	return i.dataCenterId
}
