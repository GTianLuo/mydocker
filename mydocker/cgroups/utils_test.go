package cgroups

import (
	"fmt"
	"my_docker/mydocker/cgroups/subsystems"
	"my_docker/mydocker/common/pipe"
	"testing"
)

func TestFindCgroupMountpoint(t *testing.T) {
	mountpoint := subsystems.FindCgroupMountpoint("cgroup2")
	fmt.Println(mountpoint)
}

func TestPipe(t *testing.T) {
	r, w, err := pipe.NewPipe()
	if err != nil {
		panic(err)
	}
	if err = pipe.WritePipe(w, ("hello world")); err != nil {
		panic(err)
	}
	pipe.ClosePipe(w)
	msgBytes, err := pipe.ReadPipe(r)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(msgBytes))
}
