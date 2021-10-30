package main

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/kataras/golog"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
	"github.com/kataras/iris/v12/mvc"
	"go-natok-server/controller"
	"go-natok-server/core"
	"go-natok-server/service"
	"go-natok-server/support"
	"go-natok-server/timer"
	"io/ioutil"
	"net"
	"strconv"
)

// 程序入口
func main() {

	Start()

	StartWeb()
}

// Start 启动主服务
func Start() {
	var (
		appConf  = support.Conf.Server
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
		certBytes, err := ioutil.ReadFile(appConf.CertPemPath)
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

	app.Favicon("./web/static/favicon.ico")
	app.HandleDir("/static", "./web/static")
	app.RegisterView(iris.HTML("./web/view", ".html"))

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
	app.Run(iris.Addr(":"+strconv.Itoa(support.Conf.WebPort)), iris.WithCharset("UTF-8"))
}
