package dsmapper

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"natok-server/core"
	"natok-server/model"
	"natok-server/support"
	"natok-server/util"
	"time"
	xormc "xorm.io/core"
)

// Engine 定义orm引擎
var Engine *xorm.Engine

func IsNotEmpty(str string, f func() string) string {
	if str != "" {
		return f()
	}
	return ""
}

// InitDatabase 数据库初始化
func InitDatabase() {
	logrus.Info("Init database start.")
	var (
		engine *xorm.Engine
		err    error
		dbProp = support.AppConf.Natok.Db
		dbName = "natok" + IsNotEmpty(dbProp.DbSuffix, func() string {
			return "_" + dbProp.DbSuffix
		})
	)
	if dbProp.Type == core.Mysql {
		dbUrl := initMysqlDb(dbProp, dbName)
		engine, err = xorm.NewEngine("mysql", dbUrl)
	} else if dbProp.Type == core.Sqlite {
		engine, err = xorm.NewEngine("sqlite3", fmt.Sprintf("%s./web/%s.db", support.GetCurrentAbPath(), dbName))
	} else {
		logrus.Fatalf("conf.yaml in key [natok.datasource.type] only support [%s|%s]!", core.Sqlite, core.Mysql)
	}
	if err != nil {
		logrus.Fatal("Database connection failed: ", err)
	}
	mapper := xormc.NewPrefixMapper(xormc.SnakeMapper{}, IsNotEmpty(dbProp.TablePrefix, func() string {
		return dbProp.TablePrefix + "_"
	}))
	engine.SetMapper(mapper)

	if err := engine.Sync(new(model.NatokUser), new(model.NatokClient), new(model.NatokPort), new(model.NatokTag)); err != nil {
		logrus.Fatal("Database table synchronization failed: ", err)
	}
	if count, err := engine.Count(new(model.NatokUser)); err != nil {
		logrus.Fatal("Query user record failed: ", err)
	} else if count <= 0 {
		password, err := util.GeneratePassword(16)
		if err != nil {
			panic(err)
		}
		logrus.Info("#####################################")
		logrus.Infof("Generated password: %s", password)
		logrus.Info("#####################################")
		if _, err := engine.Insert(&model.NatokUser{Username: "admin", Password: util.Md5(password)}); err != nil {
			logrus.Fatal("Failed to initialize system account.", err)
		}
	}

	engine.ShowSQL(true)
	engine.ShowExecTime(true)
	engine.SetMaxIdleConns(2)
	engine.SetMaxOpenConns(100)
	engine.SetConnMaxLifetime(time.Minute)
	Engine = engine

	logrus.Info("Init database done.")
}

// initMysqlDb 初始化数据库
func initMysqlDb(prop support.DataSource, name string) string {
	dbUrl := fmt.Sprintf("%s:%s@tcp(%s:%d)/", prop.Username, prop.Password, prop.Host, prop.Port)
	dbParam := name + "?charset=utf8mb4&parseTime=true&loc=Local"

	db, err := sql.Open("mysql", dbUrl)
	if err != nil {
		logrus.Fatal("init mysql err", err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("CREATE database if NOT EXISTS %s default character set utf8mb4 collate utf8mb4_0900_ai_ci;", name))
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("USE " + name)
	if err != nil {
		logrus.Fatal("init mysql err", err)
	}
	defer db.Close()

	logrus.Infof("Loading database %s done.", name)
	return dbUrl + dbParam
}

// StateRest 状态重置
func (d *DsMapper) StateRest() {
	// 重置客户端状态
	d.ClientStateReset()
	// 停用过期端口
	d.DisableExpiredPort()
	// 加载客户端
	if clients := d.ClientFindAll(); clients != nil {
		for _, cli := range clients {
			core.ClientManage.Store(cli.AccessKey, &core.ClientBlocking{
				AccessKey: cli.AccessKey,
				Enabled:   cli.Enabled == 1,
			})
		}
	}
	// 加载端口
	if ports := d.PortFind(true); ports != nil {
		var portIds = make([]int64, 0)
		for _, port := range ports {
			port.WhitelistNilEmpty()
			if cm, ifCM := core.ClientManage.Load(port.AccessKey); cm != nil && ifCM {
				client := cm.(*core.ClientBlocking)
				if client.Enabled {
					client.PortListener.Store(port.PortSign, &core.PortMapping{
						AccessKey: port.AccessKey,
						PortSign:  port.PortSign,
						PortNum:   port.PortNum,
						Intranet:  port.Intranet,
						Protocol:  port.Protocol,
						Whitelist: port.Whitelist,
					})
					portIds = append(portIds, port.PortId)
				}
			}
		}
		d.PortUpApply(portIds...)
	}
}
