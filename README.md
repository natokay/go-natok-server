# go-natok-server


**Go 1.13 及以上（推荐）**
```shell
# 配置 GOPROXY 环境变量
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.io,direct
```

自主构建natok-server可执行程序

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
