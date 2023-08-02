package main

import (
	log "github.com/sirupsen/logrus"
	"my_docker/mydocker/cgroups"
	"my_docker/mydocker/cgroups/subsystems"
	"my_docker/mydocker/common/pipe"
	"my_docker/mydocker/container"
	"os"
	"strconv"
)

// Run 启动容器
func Run(isTty bool, isInteractive bool, command string, res *subsystems.ResourceConfig, volume []string) {
	cmd, writePipe, err := container.NewParentProcess(isTty, isInteractive, command, volume)
	defer func() {
		rootUrl := "/home/gtl/docker"
		mntUrl := "/home/gtl/docker/mnt"
		container.DeleteWorkSpace(rootUrl, mntUrl, volume)
	}()
	if err != nil {
		log.Errorf("failed run container:%v", err)
		return
	}
	if cmd == nil {
		log.Error("failed start container")
		return
	}
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
	// 传递参数
	if err := pipe.WritePipe(writePipe, command); err != nil {
		log.Error(err)
	}
	_ = pipe.ClosePipe(writePipe)
	cmd.Wait()
	return
}
