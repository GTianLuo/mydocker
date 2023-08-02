package main

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"my_docker/mydocker/cgroups/subsystems"
	"os"
	"path/filepath"
)

func initDocker() error {

	// 初始化日志配置
	log.SetOutput(os.Stdout)
	// 初始化docker日志环境
	cgroupEnvPath := subsystems.FindCgroupMountpoint("")
	if _, err := os.Stat(cgroupEnvPath); err != nil && os.IsNotExist(err) {
		if err := os.Mkdir(cgroupEnvPath, 0775); err != nil {
			return err
		}
		if err := ioutil.WriteFile(filepath.Join(cgroupEnvPath, "cgroup.subtree_control"),
			[]byte("+cpu +memory +io +cpuset"), 0644); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}
