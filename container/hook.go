package container

import (
	"docker/command"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func sigHook(containerName string) {
	// 信号接收器
	sigCh := make(chan os.Signal, 2)
	// 监听kill信号
	signal.Notify(sigCh, syscall.SIGTERM)
	go func(ch chan os.Signal) {
		sig := <-sigCh
		log.Infof("receive signal %s", sig.String())
		command.StopContainer(containerName)
	}(sigCh)
}
