package dsmapper

import (
	"github.com/go-xorm/xorm"
	"github.com/kataras/golog"
)

// DsMapper struct 数据库映射操作
type DsMapper struct {
	_session  *xorm.Session //在事务与非事务直接改变
	atomicity bool          //原子性操作，事务支持
}

// getSession 获取连接会话
func (d *DsMapper) getSession() *xorm.Session {
	if nil == d._session {
		d._session = Engine.NewSession()
		d.atomicity = false
	}
	return d._session
}

// Transaction 开启事务
func (d *DsMapper) Transaction() error {
	d._session = Engine.NewSession()
	d.atomicity = true
	return d._session.Begin()
}

// Commit 提交事务
func (d *DsMapper) Commit() {
	defer func() {
		d._session.Close()
		d._session = nil
	}()
	if err := d._session.Commit(); err != nil {
		golog.Error("Commit error", err)
	}
}

// Rollback 回滚事务
func (d *DsMapper) Rollback() {
	defer func() {
		d._session.Close()
		d._session = nil
	}()
	if err := d._session.Rollback(); err != nil {
		golog.Error("Commit error", err)
	}
}
