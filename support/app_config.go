package support

import (
	"github.com/kataras/golog"
	"gopkg.in/yaml.v2"
	"os"
	"regexp"
)

var AppConf *AppConfig

// AppConfig 应用配置
type AppConfig struct {
	Natok        Natok `yaml:"natok"`
	BaseDirPath  string
	ProxyContent string
}
type Natok struct {
	Server  Server     `yaml:"server"`
	WebPort int        `yaml:"web.port"`
	Db      DataSource `yaml:"datasource"`
	Nginx   Nginx      `yaml:"nginx"`
}

// Server NATOK服务配置
type Server struct {
	InetHost string `yaml:"host"` // 服务器地址
	InetPort int    `yaml:"port"` // 服务器端口

	CertPemPath string `yaml:"cert-pem-path"` //密钥路径
	CertKeyPath string `yaml:"cert-key-path"` //证书路径
	LogFilePath string `yaml:"log-file-path"` //日志路径
}
type Nginx struct {
	Enable bool   `yaml:"enable"`
	Home   string `yaml:"home"`
}

// DataSource 源数据库配置
type DataSource struct {
	Host        string `yaml:"host"`         //主机
	Port        int    `yaml:"port"`         //端口
	Username    string `yaml:"username"`     //用户名
	Password    string `yaml:"password"`     //密码
	DbPrefix    string `yaml:"db-prefix"`    //数据库前綴
	TablePrefix string `yaml:"table-prefix"` //表前缀
}

// 初始化配置
func init() {
	baseDirPath := getCurrentAbPath()
	// 读取文件内容
	file, err := os.ReadFile(baseDirPath + "conf.yaml")
	if err != nil {
		panic(err)
	}
	// 利用json转换为AppConfig
	appConfig := new(AppConfig)
	err = yaml.Unmarshal(file, appConfig)
	if err != nil {
		panic(err)
	}
	appConfig.BaseDirPath = baseDirPath
	server := &appConfig.Natok.Server
	compile, err := regexp.Compile("^/|^\\\\|^[a-zA-Z]:")
	// 密钥文件
	if server.CertKeyPath != "" && !compile.MatchString(server.CertKeyPath) {
		golog.Infof("%s -> %s", server.CertKeyPath, baseDirPath+server.CertKeyPath)
		server.CertKeyPath = baseDirPath + server.CertKeyPath
	}
	// 证书文件
	if server.CertPemPath != "" && !compile.MatchString(server.CertPemPath) {
		golog.Infof("%s -> %s", server.CertPemPath, baseDirPath+server.CertPemPath)
		server.CertPemPath = baseDirPath + server.CertPemPath
	}
	// 日志文件
	if server.LogFilePath != "" && !compile.MatchString(server.LogFilePath) {
		golog.Infof("%s -> %s", server.LogFilePath, baseDirPath+server.LogFilePath)
		server.LogFilePath = baseDirPath + server.LogFilePath
	}

	AppConf = appConfig
}
