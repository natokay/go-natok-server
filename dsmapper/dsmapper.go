package dsmapper

import (
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"sync/atomic"
)

// DsMapper struct 数据库映射操作
type DsMapper struct {
	_session *xorm.Session //会话
	counter  int64         //计数
}

// getSession 获取连接会话
func (d *DsMapper) getSession() *xorm.Session {
	if nil == d._session {
		d._session = Engine.NewSession()
	}
	return d._session
}

// Transaction 事务保证一致性
func (d *DsMapper) Transaction(exec func() error) error {
	if err := d.transaction(); err != nil {
		return err
	}
	err := exec()
	if nil == err {
		d.commit()
	} else {
		d.rollback()
	}
	return err
}

// transaction 开启事务
func (d *DsMapper) transaction() error {
	if atomic.AddInt64(&d.counter, 1) == 1 {
		d._session = Engine.NewSession()
		return d._session.Begin()
	}
	return nil
}

// commit 提交事务
func (d *DsMapper) commit() {
	if atomic.AddInt64(&d.counter, -1) == 0 {
		defer func() {
			d._session.Close()
			d._session = nil
		}()
		if err := d._session.Commit(); err != nil {
			logrus.Error("commit error", err)
		}
	}
}

// rollback 回滚事务
func (d *DsMapper) rollback() {
	atomic.StoreInt64(&d.counter, 0)
	defer func() {
		d._session.Close()
		d._session = nil
	}()
	if err := d._session.Rollback(); err != nil {
		logrus.Error("rollback error", err)
	}
}
