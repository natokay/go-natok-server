package support

import (
	"encoding/json"
	"io/ioutil"
)

var Conf *AppConfig

type AppConfig struct {
	Server  Server     `json:"natok.server"`
	WebPort int        `json:"natok.web.port"`
	Db      DataSource `json:"datasource"`
	Nginx   Nginx      `json:"nginx"`

	ProxyContent string
}
type Server struct {
	InetHost    string `json:"host"`
	InetPort    int    `json:"port"`
	CertPemPath string `json:"cert-pem-path"`
	CertKeyPath string `json:"cert-key-path"`
}
type Nginx struct {
	Enable bool   `json:"enable"`
	Home   string `json:"home"`
}

// DataSource 源数据库配置
type DataSource struct {
	Host        string `json:"host"`         //主机
	Port        int    `json:"port"`         //端口
	Username    string `json:"username"`     //用户名
	Password    string `json:"password"`     //密码
	DbPrefix    string `json:"db-prefix"`    //数据库前綴
	TablePrefix string `json:"table-prefix"` //表前缀
}

// 初始化配置
func init() {
	// 读取文件内容
	file, err := ioutil.ReadFile("application.json")
	if err != nil {
		panic(err)
	}
	// 利用json转换为AppConfig
	appConfig := new(AppConfig)
	err = json.Unmarshal(file, appConfig)
	if err != nil {
		panic(err)
	}
	Conf = appConfig
}
