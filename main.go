package main

import (
	"github.com/kardianos/service"
	"github.com/sirupsen/logrus"
	"natok-server/bootstrap"
	"os"
)

type Program struct{}

func (p *Program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *Program) run() {
	logrus.Info("Started natok-server service")
	bootstrap.StartApp()
}

func (p *Program) Stop(s service.Service) error {
	logrus.Info("Stop natok-server service")
	return nil
}

// 程序入口
func main() {
	svcConfig := &service.Config{
		Name:        "natok-server",
		DisplayName: "natok-server Service",
		Description: "Go语言实现的内网代理服务端服务",
	}

	prg := &Program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			if se := s.Install(); se != nil {
				logrus.Error("Service installation failed. ", se)
			} else {
				logrus.Info("Service installed")
			}
		case "uninstall":
			if se := s.Uninstall(); se != nil {
				logrus.Error("Service uninstall failed. ", se)
			} else {
				logrus.Info("Service uninstalled")
			}
		case "start":
			if se := s.Start(); se != nil {
				logrus.Error("Service start failed. ", se)
			} else {
				logrus.Info("Service startup completed")
			}
		case "restart":
			if se := s.Restart(); se != nil {
				logrus.Error("Service restart failed. ", se)
			} else {
				logrus.Info("Service restart completed")
			}
		case "stop":
			if se := s.Stop(); se != nil {
				logrus.Error("Service stop failed. ", se)
			} else {
				logrus.Info("Service stop completed")
			}
		default:
			logrus.Warn("Unknown command: ", os.Args[1])
		}
		return
	}
	if err = s.Run(); err != nil {
		logrus.Fatal(err)
	}
}
