package model

import (
	"math"
	"time"
)

// NatokPort struct 端口对象
type NatokPort struct {
	PortId    int64     `json:"portId" xorm:"'port_id' autoincr pk not null"`          //端口主键
	AccessKey string    `json:"accessKey" xorm:"'access_key' default ''"`              //客户端秘钥
	PortSign  string    `json:"portSign" xorm:"'port_sign' default ''"`                //端口签名
	PortScope string    `json:"portScope" xorm:"'port_scope' default 'global'"`        //监听范围：global=全局,local=本地
	PortNum   int       `json:"portNum" xorm:"'port_num' default '0'"`                 //访问端口：0.0.0.0:80,127.0.0.1:80
	Intranet  string    `json:"intranet" xorm:"'intranet' default ''"`                 //转发地址：127.0.0.1:80
	Protocol  string    `json:"protocol" xorm:"'protocol' default ''"`                 //转发类型
	Whitelist []string  `json:"whitelist" xorm:"'whitelist' default null"`             //开放名单
	Tag       []int64   `json:"tag" xorm:"'tag' json default null"`                    //端口标签
	Remark    string    `json:"remark" xorm:"'remark' default ''"`                     //端口备注
	ExpireAt  time.Time `json:"expireAt" xorm:"'expire_at'"`                           //过期时间
	CreateAt  time.Time `json:"createAt" xorm:"'create_at' default CURRENT_TIMESTAMP"` //创建时间
	State     int8      `json:"state" xorm:"'state' default '0'"`                      //客户端状态：1-启动,0-停用
	Enabled   int8      `json:"enabled" xorm:"'enabled' default '0'"`                  //端口状态：1-启动,0-停用
	Apply     int8      `json:"apply" xorm:"'apply' default '0'"`                      //应用：1-已应用,0-未应用
	Modified  time.Time `json:"modified" xorm:"'modified'"`                            //修改时间
	Deleted   int8      `json:"Deleted" xorm:"'deleted' default '0'"`                  //标记移除：1-已删除,0-未删除

	ValidDay float64 `json:"validDay" xorm:"-"` // 剩余有效时间
}

// ValidDayCalculate 计算剩余有效时间
func (n *NatokPort) ValidDayCalculate() {
	validDay := float64(n.ExpireAt.Unix() - time.Now().Unix())
	if validDay > 0 && validDay < float64(86400) {
		// 不足1天，显示为1天内剩余的百分比
		validDay = math.Floor(validDay/60/60/24*100) / 100
	} else {
		// 大于1天，按天显示
		validDay = math.Floor(validDay / float64(86400))
	}
	n.ValidDay = validDay
}

// WhitelistNilEmpty 开放名单nil转[]
func (n *NatokPort) WhitelistNilEmpty() {
	if n.Whitelist == nil {
		n.Whitelist = []string{}
	}
}

// TagNilEmpty 端口标签nil转[]
func (n *NatokPort) TagNilEmpty() {
	if n.Tag == nil {
		n.Tag = []int64{}
	}
}
