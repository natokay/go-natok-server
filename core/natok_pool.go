package core

import (
	"github.com/sirupsen/logrus"
	"natok-server/util"
	"sync/atomic"
	"time"
)

// ConnectPool 连接池结构体
type ConnectPool struct {
	shutdownChan chan struct{}        // 关闭检查协程的信号
	natokHandler *ConnectHandler      // 客户端连接句柄
	natokChan    chan *ConnectHandler // 可复用连接的通道
	minSize      int                  // 最小连接数
	maxSize      int                  // 最大连接数
	current      int32                // 当前连接数
	idleTimeout  time.Duration        // 连接空闲时间
}

// NewConnectionPool 初始化连接池
func NewConnectionPool(minSize, maxSize int, idleTimeout time.Duration, connect *ConnectHandler) *ConnectPool {
	p := &ConnectPool{
		natokChan:    make(chan *ConnectHandler, maxSize),
		natokHandler: connect,
		maxSize:      maxSize,
		minSize:      minSize,
		current:      int32(0),
		idleTimeout:  idleTimeout,
		shutdownChan: make(chan struct{}),
	}
	// 初始化最小连接数
	go p.initiate(minSize)
	// 启动空闲连接清理协程
	go p.cleanIdle()
	return p
}

// Get 取出连接
func (p *ConnectPool) Get() *ConnectHandler {
	// 如果连接池中的 就绪连接数 == 最小连接数*0.4，则尝试扩容
	if factor := int(float32(p.minSize) * 0.4); len(p.natokChan) == factor {
		expand := min(p.minSize-factor+len(p.natokChan), p.maxSize-int(p.current))
		if expand > 0 {
			go p.initiate(expand)
		}
	}
	// 连接池为空，则创建新的连接
	if len(p.natokChan) == 0 {
		go p.initiate(2)
	}
	select {
	// 从连接池中获取连接
	case handler := <-p.natokChan:
		p.increment()
		logrus.Debugf("Get from connect %s, ready chan %d, active chan %d",
			handler.MsgHandler.(*NatokConnectHandler).Serial, len(p.natokChan), p.current,
		)
		return handler
	// 等待15秒，如果连接池为空则返回nil
	case <-time.After(time.Second * 15):
		return nil
	}
}

// Put 归还连接
func (p *ConnectPool) Put(handler *ConnectHandler) {
	p.decrement()
	select {
	// 成功归还连接
	case p.natokChan <- handler:
		logrus.Debugf("Put connect %s, ready chan %d, active chan %d",
			handler.MsgHandler.(*NatokConnectHandler).Serial, len(p.natokChan), p.current,
		)
	// 池满了，丢弃连接
	default:
		handler.MsgHandler.Close(handler)
		logrus.Errorf("Disconnect %s ,ready chan is full %d,active chan %d",
			handler.MsgHandler.(*NatokConnectHandler).Serial, len(p.natokChan), p.current,
		)
	}
}

// Accept 接入连接
func (p *ConnectPool) Accept(handler *ConnectHandler) {
	select {
	// 成功接入连接
	case p.natokChan <- handler:
		logrus.Debugf("Accept connect %s, ready chan %d, active chan %d",
			handler.MsgHandler.(*NatokConnectHandler).Serial, len(p.natokChan), p.current,
		)
	// 池满了，丢弃连接
	default:
		handler.MsgHandler.Close(handler)
		logrus.Errorf("Disconnect %s ,ready chan is full %d,active chan %d",
			handler.MsgHandler.(*NatokConnectHandler).Serial, len(p.natokChan), p.current,
		)
	}
}

// Shutdown 关闭连接池并停止清理
func (p *ConnectPool) Shutdown() {
	close(p.shutdownChan)
}

// cleanIdle 清理空闲连接并保持最小存活数
func (p *ConnectPool) cleanIdle() {
	// 每分钟检查一次连接池中的空闲连接
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if p.minSize >= len(p.natokChan) {
				break
			}
			now := time.Now()
			// 检查空闲连接是否超过超时时间，并进行清理
			for count := len(p.natokChan); count > p.minSize; count-- {
				select {
				case handler := <-p.natokChan:
					if now.Sub(handler.WriteTime) > p.idleTimeout {
						// 连接空闲超过超时时间，关闭并减少池中的连接数量
						handler.MsgHandler.Close(handler)
						logrus.Debugf("Idle close connect %s, ready chan %d, active chan %d",
							handler.MsgHandler.(*NatokConnectHandler).Serial, len(p.natokChan), p.current,
						)
					} else {
						// 连接仍然活跃，放回连接池
						p.natokChan <- handler
					}
				default:
					break
				}
			}
		case <-p.shutdownChan:
			close(p.natokChan)
			for handler := range p.natokChan {
				handler.MsgHandler.Close(handler)
				logrus.Debugf("Shutdown close connect %s, ready chan %d",
					handler.MsgHandler.(*NatokConnectHandler).Serial, len(p.natokChan),
				)
			}
			return
		}
	}
}

// initiate 初始化连接
func (p *ConnectPool) initiate(size int) {
	for i := 0; i < size; i++ {
		msg := Message{Type: TypeConnectNatok, Serial: util.GenerateCode(Empty), Data: []byte("initiate")}
		p.natokHandler.Write(msg)
	}
}

// increment 增加计数
func (p *ConnectPool) increment() {
	atomic.AddInt32(&p.current, +1)
}

// decrement 减少计数
func (p *ConnectPool) decrement() {
	atomic.AddInt32(&p.current, -1)
}
