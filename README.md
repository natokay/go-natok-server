# NATOK · ![GitHub Repo stars](https://img.shields.io/github/stars/natokay/go-natok-server) ![GitHub Repo stars](https://img.shields.io/github/stars/natokay/go-natok-cli)

<div align="center">
  <!-- Snake Code Contribution Map 贪吃蛇代码贡献图 -->
  <img src="grid-snake.svg" />
</div>
<p/>


- 🌱 natok是一个将局域网内个人服务代理到公网可访问的内网穿透工具。基于tcp协议、支持udp协议, 支持任何tcp上层协议（列如: http、https、ssh、telnet、data base、remote desktop....）。
- 🤔 目前市面上提供类似服务的有: 花生壳、natapp、ngrok等等。当然, 这些工具都很优秀; 但是免费提供的服务都很有限, 想要有比较好的体验都需要支付一定的套餐费用, 由于数据包会流经第三方, 因此总归有些不太友好。
- ⚡ natok-server与natok-cli都基于GO语言开发, 先天并发支持; 运行时的内存开销也很低, 一般在二十M左右。


运行natok-server相关的准备
- 公网ip的服务器主机，配置无特殊要求，当然带宽高点也好。
- 数据库：推荐sqlite，便捷无需任何配置；支持mysql，便于数据维护。

**一、natok-server使用sqlite：conf.yaml**
```yaml
natok:
  web.port: 1000 #natok·admin管理后台web页面
  server:
    port: 1001    #natok-cli的通信；若更换需与natok-cli的端口保持一致
    cert-pem-path: web/s-cert.pem #TSL加密密钥；若更换需与natok-cli保持一致
    cert-key-path: web/s-cert.key #TSL加密证书；若更换需与natok-cli保持一致
    log-file-path: web/out.log    #程序日志输出文件
  datasource:
    type: sqlite
    db-suffix: beta    #库后缀，可指定
    table-prefix: ""   #表前缀，可指定
```

**二、natok-server使用mysql：conf.yaml**
```yaml
natok:
  web.port: 1000 #natok·admin管理后台web页面
  server:
    port: 1001    #natok-cli的通信；若更换需与natok-cli的端口保持一致
    cert-pem-path: web/s-cert.pem #TSL加密密钥；若更换需与natok-cli保持一致
    cert-key-path: web/s-cert.key #TSL加密证书；若更换需与natok-cli保持一致
    log-file-path: web/out.log    #程序日志输出文件
  datasource:
    type: mysql
    host: 127.0.0.1    #自己的数据库地址
    port: 3306         #自己的数据库端口
    username: natok    #数据库账号
    password: "123456" #数据库密码
    db-suffix: beta    #库后缀，可指定
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

---

### natok-server开发环境搭建

**Go 1.22.0 及以上（推荐）**
```shell
# 配置 GOPROXY 环境变量
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
```

```shell
# 克隆项目
git clone https://github.com/natokay/go-natok-server.git

# 进入项目目录
cd go-natok-server

# 更新/下载依赖
go mod tidy
go mod vendor

# 设置目标可执行程序操作系统构架，包括 386，amd64，arm
go env -w GOARCH=amd64

# 设置可执行程序运行操作系统，支持 darwin，freebsd，linux，windows
go env -w GOOS=windows

# golang windows 程序获取管理员权限(UAC)
# go install github.com/akavel/rsrc@latest
# go env GOPATH 将里路径bin的目录配置到环境变量
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

**natok:1.3.0**
natok-cli与natok-server可支持udp网络代理。

**natok:1.4.0**
natok-server端口访问支持白名单限制，重要端口(如：linux-22,windows-3389)可限制访问的ip地址。

**natok:1.5.0**
natok-server数据库类型支持sqlite、mysql，推荐使用sqlite，部署更便捷。

**natok:1.6.0**
natok-server与natok-client内部通讯采用连接池，即从公网访问natok-server后，会将连接放入连接池中，以便后续的请求时能更快的响应。

**natok:1.6.1**
natok-server的访问端口监听，可选择监听范围：global=全局,local=本地。


## NATOK平台界面预览

登录页面
![image-20250303220714-r1kbi0b](https://github.com/user-attachments/assets/49e963e1-0062-4e2b-89d2-8309472e9fe7)

统计概览
![image-20250303220743-etmceyf](https://github.com/user-attachments/assets/cba87be9-e6d0-4ab2-8fbe-222397c4a06a)

代理管理
![image-20250303220953-vz1hjpb](https://github.com/user-attachments/assets/bc42a243-c1fc-4fa3-adfd-23c6175f9166)
![image-20250303221323-a0q00lk](https://github.com/user-attachments/assets/ff38b0a3-d578-4342-a68c-98e4775c5021)

端口映射
![image-20250303221053-j7b3tsy](https://github.com/user-attachments/assets/4f65aea5-5f97-42dc-94a0-0e3af73d4bef)
![image-20250303221456-pkfl4wt](https://github.com/user-attachments/assets/3692fce0-6104-47ee-b2b5-fcafd78366ec)

标签名单
![image-20250303221123-zl9f76j](https://github.com/user-attachments/assets/02262934-f260-43da-8435-45fdd35c1793)
![image-20250303221545-9n2vwqs](https://github.com/user-attachments/assets/14ddd49a-fdcc-49d0-ae8e-071a9962ac4c)
