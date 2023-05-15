package main

import (
	"crypto/tls"
	"crypto/x509"
	win "github.com/kardianos/service"
	"github.com/kataras/golog"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
	"github.com/kataras/iris/v12/mvc"
	"natok-server/controller"
	"natok-server/core"
	"natok-server/service"
	"natok-server/support"
	"natok-server/timer"
	"net"
	"os"
	"strconv"
)

type Program struct{}

func (p *Program) Start(s win.Service) error {
	go p.run()
	return nil
}

func (p *Program) run() {
	golog.Info("Started natok server service")
	Start()
	StartWeb()
}

func (p *Program) Stop(s win.Service) error {
	golog.Info("Stop natok server service")
	return nil
}

func init() {
	// 日志记录处理
	golog.SetLevel("debug")
	logFilePath := support.AppConf.Natok.Server.LogFilePath
	if logFilePath != "" {
		logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			golog.Error(err)
		} else {
			golog.AddOutput(logFile)
		}
	}
}

// 程序入口
func main() {
	svcConfig := &win.Config{
		Name:        "natok-server",
		DisplayName: "Natok Server Service",
		Description: "Go语言实现的内网代理服务端服务",
	}

	prg := &Program{}
	s, err := win.New(prg, svcConfig)
	if err != nil {
		golog.Fatal(err)
	}

	if len(os.Args) > 1 {
		if os.Args[1] == "install" {
			if se := s.Install(); se != nil {
				golog.Error("Service installation failed. ", se)
			} else {
				golog.Info("Service installed")
			}
			return
		}
		if os.Args[1] == "uninstall" {
			if se := s.Uninstall(); se != nil {
				golog.Error("Service uninstall failed. ", se)
			} else {
				golog.Info("Service uninstalled")
			}
			return
		}
		if os.Args[1] == "start" {
			if se := s.Start(); se != nil {
				golog.Error("Service start failed. ", se)
			} else {
				golog.Info("Service startup completed")
			}
			return
		}
		if os.Args[1] == "restart" {
			if se := s.Restart(); se != nil {
				golog.Error("Service restart failed. ", se)
			} else {
				golog.Info("Service restart completed")
			}
			return
		}
		if os.Args[1] == "stop" {
			if se := s.Stop(); se != nil {
				golog.Error("Service stop failed. ", se)
			} else {
				golog.Info("Service stop completed")
			}
			return
		}
	}

	if err = s.Run(); err != nil {
		golog.Fatal(err)
	}
}

// Start 启动主服务
func Start() {
	var (
		appConf  = support.AppConf.Natok.Server
		addr     = appConf.InetHost + ":" + strconv.Itoa(appConf.InetPort)
		listener net.Listener
		err      error
	)

	if appConf.CertKeyPath == "" || appConf.CertPemPath == "" {
		listener, err = net.Listen("tcp", addr)
		golog.Info("NET Listen")
	} else {
		cert, err := tls.LoadX509KeyPair(appConf.CertPemPath, appConf.CertKeyPath)
		if err != nil {
			golog.Fatal(err)
		}
		certBytes, err := os.ReadFile(appConf.CertPemPath)
		if err != nil {
			panic("Unable to read cert.pem")
		}
		clientCertPool := x509.NewCertPool()
		ok := clientCertPool.AppendCertsFromPEM(certBytes)
		if !ok {
			panic("Failed to parse root certificate")
		}
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    clientCertPool,
		}
		listener, err = tls.Listen("tcp", addr, tlsConfig)
		golog.Info("TLS Listen")
	}

	if err != nil {
		golog.Fatal(err)
	}
	golog.Info("Startup natok server on ", listener.Addr())

	// 监听来自Natok-cli的启动连接
	go func(listener net.Listener) {
		for {
			accept, err := listener.Accept()
			if err != nil {
				golog.Error("Listen failed! ", listener.Addr(), err)
				continue
			}
			go func(conn net.Conn) {
				handler := core.ConnectHandler{}
				handler.Listen(conn, &core.NatokServerHandler{
					ConnHandler: &handler,
				})
			}(accept)
		}
	}(listener)

	// 周期性任务
	timer.Worker()
}

// StartWeb 启动Web服务
func StartWeb() {
	app := iris.New()
	app.Logger().SetLevel("info")
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(support.CorsHandler())
	app.Use(support.AuthorHandler())

	baseDirPath := support.AppConf.BaseDirPath

	app.Favicon(baseDirPath + "./web/static/favicon.ico")
	app.HandleDir("/static", baseDirPath+"./web/static")
	app.RegisterView(iris.HTML(baseDirPath+"./web/view", ".html"))

	// 跨域访问配置
	visitApp := app.Party("/").AllowMethods(iris.MethodOptions)
	// MVC控制层配置
	mvc.Configure(visitApp, func(app *mvc.Application) {
		app.Register(support.SessionsManager.Start).Handle(new(controller.AuthController))
		app.Register(new(service.ClientService)).Handle(new(controller.ClientController))
		app.Register(new(service.PortService)).Handle(new(controller.PortController))
		app.Register(new(service.ReportService)).Handle(new(controller.ReportController))
	})
	// 路由首页
	visitApp.Get("/", func(ctx iris.Context) {
		ctx.Redirect("/index.html", iris.StatusFound)
	})
	visitApp.Get("/index.html", func(ctx iris.Context) {
		ctx.View("index.html")
	})
	// 启动服务，端口监听
	app.Run(iris.Addr(":"+strconv.Itoa(support.AppConf.Natok.WebPort)), iris.WithCharset("UTF-8"))
}
