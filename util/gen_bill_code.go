package util

import (
	"fmt"
	"sync"
	"time"
)

// Encoder 编码生成器
type Encoder struct {
	prefix       string
	serialNumber int
	systemSerial string
	mu           sync.Mutex
}

// Generate 生成编码
func (e *Encoder) Generate() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.serialNumber >= 9999 {
		time.Sleep(10 * time.Millisecond)
	}
	currentSerial := time.Now().Format("20060102150405")
	if currentSerial != e.systemSerial {
		e.systemSerial = currentSerial
		e.serialNumber = 1
	} else {
		e.serialNumber++
	}
	return fmt.Sprintf("%s%s%04d", e.prefix, e.systemSerial, e.serialNumber)
}

// NewEncoder 创建一个新的编码生成器
func NewEncoder(prefix string) *Encoder {
	return &Encoder{
		prefix:       prefix,
		serialNumber: 0,
		systemSerial: "",
	}
}

var (
	encoderMap sync.Map
	mu         sync.Mutex
)

// GenerateCode 生成编码
func GenerateCode(prefix string) string {
	if encoder, ok := encoderMap.Load(prefix); encoder != nil && ok {
		return encoder.(*Encoder).Generate()
	}
	mu.Lock()
	defer mu.Unlock()
	encoder := NewEncoder(prefix)
	encoderMap.Store(prefix, encoder)
	return encoder.Generate()
}
