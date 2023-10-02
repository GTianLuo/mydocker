package subsystems

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"testing"
)

//func TestCgroupManager(t *testing.T) {
//	res := &ResourceConfig{
//		MemoryLimit: "200M",
//	}
//	cgroupManager := cgroups.NewCgroupManager("/container-"+strconv.Itoa(os.Getpid()), res)
//	defer cgroupManager.Destroy()
//	if err := cgroupManager.Apply(os.Getpid()); err != nil {
//		panic(err)
//	}
//	if err := cgroupManager.Set(); err != nil {
//		panic(err)
//	}
//}

func TestExec(t *testing.T) {
	command := "ls"
	path, err := exec.LookPath("ls")
	if err != nil {
		panic(err)
	}
	err = syscall.Exec(path, []string{command}, os.Environ())
	fmt.Println(err)
}

func TestPwd(t *testing.T) {
	fmt.Println(os.Getwd())
}
