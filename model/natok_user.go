package model

// NatokUser struct 用户对象
type NatokUser struct {
	Id       int64  `json:"id" xorm:"'id' autoincr pk notnull"` //索引
	Nick     string `json:"nick" xorm:"'nick'"`
	Username string `json:"username" xorm:"'username' notnull"` //用户名最小长度不能少于3个字符，最大超度不得超过128个字符
	Password string `json:"password" xorm:"'password' notnull"` //密码最小长度不能少于6个字符，最大超度不得超过128个字符
	Email    string `json:"email" xorm:"'email'"`               //邮箱
	Phone    string `json:"phone" xorm:"'phone'"`               //电话
	Token    string `json:"token" xorm:"'token'"`               //访问令牌

	Code string `json:"code" xorm:"-"` //验证码
}
