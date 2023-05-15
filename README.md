# NATOK

- natok是一个将局域网内个人服务代理到公网可访问的内网穿透工具，基于tcp协议，支持任何tcp上层协议（列如：http、https、ssh、telnet、data base、remote desktop....）。
- 目前市面上提供类似服务的有：花生壳、natapp、ngrok等等。当然，这些工具都很优秀！但是免费提供的服务都很有限，想要有比较好的体验都需要支付一定的套餐费用，由于数据包会流经第三方，因此总归有些不太友好。
- natok-server与natok-cli都基于GO语言开发，几乎不存在并发问题。运行时的内存开销也很低，一般在几十M左右。所以很推荐自主搭建服务！


**服务端与客户端**

| 服务                     |支持系统| 下载地址                                               |
| ------------------------|----- | ------------------------------------------------------ |
| natok-cli |linux/windows| [GitHub](https://github.com/natokay/go-natok-cli/releases) |
| natok-server| linux/windows|[GitHub](https://github.com/natokay/go-natok-server/releases) |

# go-natok-server:1.2.0


运行natok-server相关的准备
- 公网ip的服务器主机，配置无特殊要求，当然带宽高点也好
- 服务器主机可访问的mysql数据库，现在的docker已经很方便了

natok-server的相关配置：application.json
```yaml
natok:
  web.port: 1000 #natok·admin管理后台web页面端口，可自定义
  server:
    host: 0.0.0.0 #natok-server与服务器地址邦定，不推荐更改
    port: 1001    #natok-cli的通信端口，可自定义。注：需与natok-cli配置文件同步
    cert-pem-path: web/s-cert.pem #TSL加密密钥，可自己指定。注：需与natok-cli端保持一致
    cert-key-path: web/s-cert.key #TSL加密证书，可自己指定。注：需与natok-cli端保持一致
    log-file-path: web/out.log    #程序日志输出配置
  datasource: #Mysql数据源配置
    host: 127.0.0.1    #自己的数据库地址
    port: 3306         #自己的数据库端口
    username: natok    #数据库用户名
    password: "123456" #数据库密码
    db-prefix: playxy  #数据库前缀，可指定
    table-prefix: ""   #表前缀，可指定
```

- windows系统启动： 双击 natok-server.exe
```powershell
# 注册服务，自动提取管理员权限：
natok-server.exe install
# 卸载服务，自动提取管理员权限：
natok-server.exe uninstall
# 启停服务，自动提取管理员权限：
natok-server.exe start/stop
# 启停服务，终端管理员权限
net start/stop natok-server
```
- Linux系统启动：
```shell
# 授予natok-server可执权限
chmod 755 natok-server
# 启动应用
nohup ./natok-server > /dev/null 2>&1 &
```

**Go 1.13 及以上（推荐）**
```shell
# 配置 GOPROXY 环境变量
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
```

构建natok-server可执行程序

```shell
# 克隆项目
git clone https://github.com/natokay/go-natok-server.git

# 进入项目目录
cd go-natok-server

# 更新/下载依赖
go mod tidy
go mod vendor

# 设置目标可执行程序操作系统构架，包括 386，amd64，arm
set GOARCH=amd64

# 设置可执行程序运行操作系统，支持 darwin，freebsd，linux，windows
set GOOS=windows

# golang windows 程序获取管理员权限(UAC)
rsrc -manifest nac.manifest -o nac.syso

# cd到main.go目录，打包命令
go build

# 启动程序
./natok-server.exe
```

## 版本描述
**natok:1.0.0**
natok-cli与natok-server网络代理通信基本功能实现。

**natok:1.1.0**
natok-cli与natok-server支持windows平台注册为服务运行，可支持开机自启，保证服务畅通。

**natok:1.2.0**
natok-cli可与多个natok-server保持连接，支持从多个不同的natok-server来访问natok-cli，以实现更快及更优的网络通信。
