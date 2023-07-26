package subsystems

import (
	"os"
	"strconv"
	"testing"
)

func TestMemorySubsystem(t *testing.T) {
	res := &ResourceConfig{
		MemoryLimit: "200M",
	}
	mSub := &MemorySubsystem{}
	cgroupPath := "/container-" + strconv.Itoa(os.Getpid())
	if err := mSub.Apply(cgroupPath, os.Getpid()); err != nil {
		panic(err)
	}
	if err := mSub.Set(cgroupPath, res); err != nil {
		panic(err)
	}
}
