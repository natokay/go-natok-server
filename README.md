# go-natok-server


运行natok-server相关的准备
- 公网ip的服务器主机，配置无特殊要求，当然带宽高点也好
- 服务器主机可访问的mysql数据库，现在的docker已经很方便了

natok-server的相关配置：application.json

```json5
{
  "natok.web.port": 1000,               // natok·admin管理后台web页面端口，可自定义
  "natok.server": {
    "host": "0.0.0.0",                  // natok-server与服务器地址邦定，不推荐更改
    "port": 1001,                       // natok-cli的通信端口，可自定义。注：需与natok-cli配置文件同步
    "cert-key-path": "web/s-cert.key",  // TSL加密密钥，可自己指定。注：需与natok-cli端保持一致
    "cert-pem-path": "web/s-cert.pem"   // TSL加密证书，可自己指定。注：需与natok-cli端保持一致
  },
  "datasource": {                       // Mysql数据源配置
    "host": "127.0.0.1",                // 自己的数据库地址
    "port": 3306,                       // 自己的数据库端口
    "username": "natok",                // 数据库用户名
    "password": "123456",               // 数据库密码
    "db-prefix": "",                    // 数据库前缀，可指定
    "table-prefix": ""                  // 表前缀，可指定
  }
}
```



**Go 1.13 及以上（推荐）**
```shell
# 配置 GOPROXY 环境变量
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.io,direct
```

构建natok-server可执行程序

```shell
# 克隆项目
git clone https://github.com/play-sy/go-natok-server.git

# 进入项目目录
cd go-natok-server

# 更新/下载依赖
go mod vendor

# 设置目标可执行程序操作系统构架，包括 386，amd64，arm
set GOARCH=amd64

# 设置可执行程序运行操作系统，支持 darwin，freebsd，linux，windows
set GOOS=windows

# cd到main.go目录，打包命令
go build

# 启动程序
./go-natok-server.exe
```
