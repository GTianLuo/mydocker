package main

import (
	log "github.com/sirupsen/logrus"
	"my_docker/mydocker/cgroups"
	"my_docker/mydocker/cgroups/subsystems"
	"my_docker/mydocker/container"
	"os"
	"strconv"
)

// Run 启动容器
func Run(isTty bool, isInteractive bool, command string, res *subsystems.ResourceConfig) {
	cmd := container.NewParentProcess(isTty, isInteractive, command)
	if err := cmd.Start(); err != nil {
		log.Error(err)
		return
	}
	// 使用container-pid作为cgroup名字
	cgroupManager := cgroups.NewCgroupManager("container-"+strconv.Itoa(os.Getpid()), res)
	defer func() {
		if err := cgroupManager.Destroy(); err != nil {
			log.Error("destroy container failed", err.Error())
		}
	}()

	if err := cgroupManager.Apply(cmd.Process.Pid); err != nil {
		log.Error(err)
		return
	}
	if err := cgroupManager.Set(); err != nil {
		log.Error(err)
		return
	}
	cmd.Wait()
	return
}
