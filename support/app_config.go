package support

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io"
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
	Debug   bool       `yaml:"log.debug"`
	Db      DataSource `yaml:"datasource"`
}

// Server NATOK服务配置
type Server struct {
	InetHost    string   `yaml:"host"`          //服务器地址
	InetPort    int      `yaml:"port"`          //服务器端口
	LogFilePath string   `yaml:"log-file-path"` //日志路径
	CertPemPath string   `yaml:"cert-pem-path"` //密钥路径
	CertKeyPath string   `yaml:"cert-key-path"` //证书路径
	ChanPool    ChanPool `yaml:"chan-pool"`     //连接池配置
}

// ChanPool 通道连接池配置
type ChanPool struct {
	MaxSize     int   `yaml:"max-size"`     //最大连接数
	MinSize     int   `yaml:"min-size"`     //最小连接数
	IdleTimeout int64 `yaml:"idle-timeout"` //连接空闲时间(秒)
}

// DataSource 源数据库配置
type DataSource struct {
	Type        string `yaml:"type"`         //数据类型：sqlite、mysql
	Host        string `yaml:"host"`         //主机
	Port        int    `yaml:"port"`         //端口
	Username    string `yaml:"username"`     //用户名
	Password    string `yaml:"password"`     //密码
	DbSuffix    string `yaml:"db-suffix"`    //库后缀
	TablePrefix string `yaml:"table-prefix"` //表前缀
}

// SetDefaults 设置默认值
func (c *ChanPool) SetDefaults() {
	if c.MaxSize == 0 {
		c.MaxSize = 200
	}
	if c.MinSize == 0 {
		c.MinSize = 10
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = 600
	}
}

// InitConfig 初始化配置
func InitConfig() {
	if AppConf != nil {
		return
	}
	baseDirPath := GetCurrentAbPath()
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
	// 设置默认值
	appConfig.Natok.Server.ChanPool.SetDefaults()

	// 配置当前路径
	appConfig.BaseDirPath = baseDirPath
	server := &appConfig.Natok.Server
	compile, err := regexp.Compile("^/|^\\\\|^[a-zA-Z]:")
	// 密钥文件
	if server.CertKeyPath != "" && !compile.MatchString(server.CertKeyPath) {
		logrus.Infof("%s -> %s", server.CertKeyPath, baseDirPath+server.CertKeyPath)
		server.CertKeyPath = baseDirPath + server.CertKeyPath
	}
	// 证书文件
	if server.CertPemPath != "" && !compile.MatchString(server.CertPemPath) {
		logrus.Infof("%s -> %s", server.CertPemPath, baseDirPath+server.CertPemPath)
		server.CertPemPath = baseDirPath + server.CertPemPath
	}
	// 日志记录
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		ForceColors:     true,
		TimestampFormat: "2006-01-02 15:04:05.000",
	})
	// 在输出日志中添加文件名和方法信息
	if appConfig.Natok.Debug {
		logrus.SetReportCaller(true)
		logrus.SetLevel(logrus.DebugLevel)
	}
	// 日志记录输出文件
	if server.LogFilePath != "" && !compile.MatchString(server.LogFilePath) {
		logrus.Infof("%s -> %s", server.LogFilePath, baseDirPath+server.LogFilePath)
		server.LogFilePath = baseDirPath + server.LogFilePath
		logFile, err := os.OpenFile(server.LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logrus.Fatal(err)
		} else {
			// 组合一下即可，os.Stdout代表标准输出流
			multiWriter := io.MultiWriter(logFile, os.Stdout)
			logrus.SetOutput(multiWriter)
		}
	}

	AppConf = appConfig
}
