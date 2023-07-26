package subsystems

import (
	"errors"
	"fmt"
	"io/ioutil"
	"my_docker/mydocker/cgroups"
	"os"
	"path"
	"strconv"
)

type MemorySubsystem struct {
}

func (m *MemorySubsystem) Name() string {
	return "memory"
}

// Set memory资源限制
func (m *MemorySubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	if subsystemCgroupPath, err := cgroups.GetCgroupPath(m.Name(), cgroupPath, true); err != nil {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	} else if res.MemoryLimit != "" {
		// memory限制写入 memory.max
		mPath := path.Join(subsystemCgroupPath, "memory.max")
		if err := ioutil.WriteFile(mPath, []byte(res.MemoryLimit), 0644); err != nil {
			return errors.New("failed set cgroup memory" + err.Error())
		}
		return nil
	}
	return nil
}

// Apply 进程应用到cgroup中
func (m *MemorySubsystem) Apply(cgroupPath string, pid int) error {
	if subsystemCgroupPath, err := cgroups.GetCgroupPath(m.Name(), cgroupPath, true); err == nil {
		// 把pid写入cgroup.proc文件
		if err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "cgroup.proc"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return errors.New("failed apply pid to this cgroup: " + err.Error())
		}
		return nil
	} else {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	}
}

// Remove 将进程移除group
// 一个容器对应一个cgroup，该操作可直接删除cgroup
func (m *MemorySubsystem) Remove(cgroupPath string) error {
	if subsystemCgroupPath, err := cgroups.GetCgroupPath(m.Name(), cgroupPath, true); err != nil {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	} else {
		err := os.Remove(subsystemCgroupPath)
		if err != nil {
			return errors.New("failed remove cgroup:" + err.Error())
		}
		return nil
	}
}
