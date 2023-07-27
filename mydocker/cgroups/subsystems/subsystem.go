package subsystems

// ResourceConfig 内存限制，CPU时间片权重，CPU核心数
type ResourceConfig struct {
	MemoryLimit string
	CpuShare    string
	CpuSet      string
}

type Subsystem interface {
	// 返回subsystem的名字
	Name() string
	Set(cgroupPath string, res *ResourceConfig) error
}
