package model

import (
	"natok-server/util"
	"time"
)

// NatokPort struct 端口对象
type NatokPort struct {
	PortId    int64     `json:"portId" xorm:"'port_id' autoincr pk notnull"` //索引
	AccessKey string    `json:"accessKey" xorm:"'access_key' default ''"`    //访问秘钥
	Sign      string    `json:"sign" xorm:"'sign' default ''"`               //映射签名
	PortNum   int       `json:"portNum" xorm:"'port_num' default '0'"`       //公网端口号123.207.217.100:80
	Intranet  string    `json:"intranet" xorm:"'intranet' default ''"`       //内网地址127.0.0.1:80
	Domain    string    `json:"domain" xorm:"'domain' default ''"`           //playxy.cn:80
	Protocol  string    `json:"protocol" xorm:"'protocol' default ''"`       //转发类型
	Remark    string    `json:"remark" xorm:"'remark' default ''"`           //端口备注
	ExpireAt  time.Time `json:"expireAt" xorm:"'expire_at'"`                 //过期时间
	CreateAt  time.Time `json:"createAt" xorm:"'create_at' default now()"`   //创建时间
	Enabled   int8      `json:"enabled" xorm:"'enabled' default '0'"`        //状态
	State     int8      `json:"state" xorm:"'state' default '0'"`            //C端状态：1:启动,0:停用
	Apply     int8      `json:"apply" xorm:"'apply' default '0'"`            //应用：0-未应用，1-已应用
	Modified  time.Time `json:"modified" xorm:"'modified'"`                  //修改时间
	Deleted   int8      `json:"Deleted" xorm:"'deleted' default '0'"`        //标记移除

	ValidDay int64 `json:"validDay" xorm:"-"`
}

func (n *NatokPort) GetValidDay() int64 {
	if n.ValidDay < 0 && (n.ExpireAt.Unix() < util.NowDayStart().Unix()) {
		return -1
	}
	n.ValidDay = (n.ExpireAt.Unix() - time.Now().Unix()) / int64(86400)
	return n.ValidDay
}

type NatokProcotol struct {
	procoltol string
	count     int
}
