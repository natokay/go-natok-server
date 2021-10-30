package model

import (
	"time"
)

// NatokClient struct 客户端对象
type NatokClient struct {
	ClientId   int64     `json:"clientId" xorm:"'client_id' autoincr pk notnull"` //客户ID
	ClientName string    `json:"clientName" xorm:"'client_name' default ''"`      //支持的用户名长度范围6-64个字符
	AccessKey  string    `json:"accessKey" xorm:"'access_key' default ''"`        //访问秘钥
	JoinTime   time.Time `json:"joinTime" xorm:"'join_time' default now()"`       //加入时间
	Enabled    int8      `json:"enabled" xorm:"'enabled' default '0'"`            //是否启用
	State      int       `json:"state" xorm:"'state' default '0'"`                //用户状态
	Apply      int8      `json:"apply" xorm:"'apply' default '0'"`                //应用：0-未应用，1-已应用
	Modified   time.Time `json:"modified" xorm:"'modified'"`                      //修改时间
	Deleted    int8      `json:"Deleted" xorm:"'deleted' default '0'"`            //标记移除

	UsePortNum int `json:"usePortNum" xorm:"-"` //使用端口数
}
