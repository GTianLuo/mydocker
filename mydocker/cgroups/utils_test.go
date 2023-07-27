package cgroups

import (
	"fmt"
	"my_docker/mydocker/cgroups/subsystems"
	"testing"
)

func TestFindCgroupMountpoint(t *testing.T) {
	mountpoint := subsystems.FindCgroupMountpoint("cgroup2")
	fmt.Println(mountpoint)
}
