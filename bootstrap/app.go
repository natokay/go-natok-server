package bootstrap

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
	"github.com/kataras/iris/v12/mvc"
	"github.com/sirupsen/logrus"
	"natok-server/controller"
	"natok-server/core"
	"natok-server/dsmapper"
	"natok-server/service"
	"natok-server/support"
	"natok-server/timer"
	"net"
	"os"
	"strconv"
)

func StartApp() {
	support.InitConfig()
	support.InitAuth()
	dsmapper.InitDatabase()
	StartServer()
	StartWeb()
}

// StartServer 启动主服务
func StartServer() {
	var (
		conf     = support.AppConf.Natok.Server
		addr     = conf.InetHost + ":" + strconv.Itoa(conf.InetPort)
		listener net.Listener
		err      error
	)

	if conf.CertKeyPath == "" || conf.CertPemPath == "" {
		listener, err = net.Listen("tcp", addr)
		logrus.Info("NET Listen")
	} else {
		cert, err := tls.LoadX509KeyPair(conf.CertPemPath, conf.CertKeyPath)
		if err != nil {
			logrus.Fatal(err)
		}
		certBytes, err := os.ReadFile(conf.CertPemPath)
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
		logrus.Info("TLS Listen")
	}

	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Infof("Startup natok-server on %s", listener.Addr())

	// 监听来自Natok-cli的启动连接
	go core.NatokClient(listener)

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
		app.Register(new(service.TagService)).Handle(new(controller.TagController))
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
