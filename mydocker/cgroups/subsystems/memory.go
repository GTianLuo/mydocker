package subsystems

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
)

type MemorySubsystem struct {
}

func (m *MemorySubsystem) Name() string {
	return "memory"
}

// Set memory资源限制
func (m *MemorySubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	if subsystemCgroupPath, err := GetCgroupPath(m.Name(), cgroupPath, true); err != nil {
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
