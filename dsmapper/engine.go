package dsmapper

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/kataras/golog"
	"natok-server/core"
	"natok-server/model"
	"natok-server/support"
	"natok-server/util"
	"time"
	xormc "xorm.io/core"
)

// 定义orm引擎
var Engine *xorm.Engine

// 创建orm引擎
func init() {
	db := support.AppConf.Natok.Db
	dbname := db.DbPrefix + "_natok"
	// 拼装数据库连接地址
	dbUrl := fmt.Sprintf("%s:%s@tcp(%s:%d)/", db.Username, db.Password, db.Host, db.Port)
	dbUrlParam := dbname + "?charset=utf8mb4&parseTime=true&loc=Local"
	initDB(dbUrl, dbname)

	engine, err := xorm.NewEngine("mysql", dbUrl+dbUrlParam)
	if err != nil {
		golog.Fatal("数据库连接失败:", err)
	}

	mapper := xormc.NewPrefixMapper(xormc.SnakeMapper{}, db.TablePrefix)
	engine.SetMapper(mapper)

	if err := engine.Sync(new(model.NatokUser), new(model.NatokClient), new(model.NatokPort)); err != nil {
		golog.Fatal("数据表同步失败:", err)
	}
	if count, err := engine.Count(new(model.NatokUser)); err != nil {
		golog.Fatal("查询用户记录表失败:", err)
	} else if count <= 0 {
		engine.Insert(&model.NatokUser{Username: "admin", Password: util.Md5("123456")})
	}

	engine.ShowSQL(true)
	engine.ShowExecTime(true)
	engine.SetMaxIdleConns(2)
	engine.SetMaxOpenConns(100)
	engine.SetConnMaxLifetime(time.Minute)
	Engine = engine
}

// initDB 初始化数据库
func initDB(dbUrl, dbName string) {
	db, err := sql.Open("mysql", dbUrl)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("CREATE database if NOT EXISTS %s default character set utf8mb4 collate utf8mb4_unicode_ci;", dbName))
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("USE " + dbName)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	golog.Info("Loading database ", dbName, " done.")
}

// Initialization 将数据库数据载入
func Initialization() {
	//TODO 重置客户端状态
	ClientResetState()
	//TODO 客户端查询
	if clients := ClientFind(true); clients != nil {
		var cliIds = make([]int64, 0)
		for _, cli := range clients {
			core.ClientGroupManage.Store(cli.AccessKey, &core.ClientBlocking{
				AccessKey: cli.AccessKey,
				Enabled:   cli.Enabled > 0,
				Mapping:   make(map[string]*core.PortMapping, 0),
			})
			cliIds = append(cliIds, cli.ClientId)
		}
		ClientUpApply(cliIds...)
	}
	//TODO 端口查询
	if ports := PortFind(true); ports != nil {
		var portIds = make([]int64, 0)
		for _, port := range ports {
			if item, ok := core.ClientGroupManage.Load(port.AccessKey); item != nil && ok {
				handler := item.(*core.ClientBlocking)
				if handler.Enabled {
					mapping := handler.Mapping
					mapping[port.Sign] = &core.PortMapping{
						AccessKey: port.AccessKey,
						Sign:      port.Sign,
						Port:      port.PortNum,
						Intranet:  port.Intranet,
						Domain:    port.Domain,
						Protocol:  port.Protocol,
					}
					portIds = append(portIds, port.PortId)
				}
			}
		}
		PortUpApply(portIds...)
	}
}
