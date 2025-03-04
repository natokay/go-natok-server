package model

import "time"

// NatokTag struct 标签对象
type NatokTag struct {
	TagId     int64     `json:"tagId" xorm:"'tag_id' autoincr pk not null"`         //标签主键
	TagName   string    `json:"tagName" xorm:"'tag_name' default ''"`               //标签名称
	Remark    string    `json:"remark" xorm:"'remark' default ''"`                  //标签备注
	Whitelist []string  `json:"whitelist" xorm:"'whitelist' default null"`          //开放名单
	Enabled   int8      `json:"enabled" xorm:"'enabled' default '0'"`               //端口状态：1-启动,0-停用
	Created   time.Time `json:"created" xorm:"'created' default CURRENT_TIMESTAMP"` //创建时间
	Modified  time.Time `json:"modified" xorm:"'modified' default null"`            //修改时间
	Deleted   int8      `json:"deleted" xorm:"'deleted' default '0'"`               //标记移除：1-已删除,0-未删除
}

// WhitelistNilEmpty 开放名单nil转[]
func (n *NatokTag) WhitelistNilEmpty() {
	if n.Whitelist == nil {
		n.Whitelist = []string{}
	}
}
