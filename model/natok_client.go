package model

import (
	"time"
)

// NatokClient struct 客户端对象
type NatokClient struct {
	ClientId   int64     `json:"clientId" xorm:"'client_id' autoincr pk notnull"`       //客户端主键
	ClientName string    `json:"clientName" xorm:"'client_name' default ''"`            //客户端名称
	AccessKey  string    `json:"accessKey" xorm:"'access_key' default ''"`              //客户端秘钥
	JoinTime   time.Time `json:"joinTime" xorm:"'join_time' default CURRENT_TIMESTAMP"` //加入时间
	Enabled    int8      `json:"enabled" xorm:"'enabled' default '0'"`                  //启用状态：1-启动,0-停用
	State      int8      `json:"state" xorm:"'state' default '0'"`                      //在线状态：1-在线,0-离线
	Modified   time.Time `json:"modified" xorm:"'modified'"`                            //修改时间
	Deleted    int8      `json:"Deleted" xorm:"'deleted' default '0'"`                  //标记移除

	UsePortNum int `json:"usePortNum" xorm:"-"` //使用端口数
}
