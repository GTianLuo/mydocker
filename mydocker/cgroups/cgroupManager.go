package cgroups

import (
	"errors"
	"fmt"
	"io/ioutil"
	"my_docker/mydocker/cgroups/subsystems"
	"os"
	"path"
	"strconv"
)

// SubsystemIns subsystem实例的处理链数组
var (
	SubsystemIns = []subsystems.Subsystem{
		&subsystems.MemorySubsystem{},
	}
)

type CgroupManager struct {
	// mydocker cgroup相对于cgroup root的路径
	path string
	// 资源限制
	Resource *subsystems.ResourceConfig
}

func NewCgroupManager(path string, res *subsystems.ResourceConfig) *CgroupManager {
	return &CgroupManager{
		path:     path,
		Resource: res,
	}
}

// Set 通过子系统链限制资源
func (m *CgroupManager) Set() error {
	for _, s := range SubsystemIns {
		if err := s.Set(m.path, m.Resource); err != nil {
			return err
		}
	}
	return nil
}

// Apply 进程应用到cgroup中
func (m *CgroupManager) Apply(pid int) error {
	if subsystemCgroupPath, err := subsystems.GetCgroupPath("", m.path, true); err == nil {
		// 把pid写入cgroup.proc文件
		if err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "cgroup.procs"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return errors.New("failed apply pid to this cgroup: " + err.Error())
		}
		return nil
	} else {
		return fmt.Errorf("get cgroup %s error: %v", m.path, err)
	}
}

// Destroy 销毁group
func (m *CgroupManager) Destroy() error {
	if subsystemCgroupPath, err := subsystems.GetCgroupPath("", m.path, true); err != nil {
		return fmt.Errorf("get cgroup %s error: %v", m.path, err)
	} else {
		err := os.Remove(subsystemCgroupPath)
		if err != nil {
			return errors.New("failed remove cgroup:" + err.Error())
		}
		return nil
	}
}
