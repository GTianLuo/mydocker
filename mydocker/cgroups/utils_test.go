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

func lengthOfLongestSubstring(s string) int {
	m := make(map[uint8]uint8, 30)
	max := 0
	l := 0
	for i := 0; i < len(s); i++ {
		if m[s[i]] == 0 {
			m[s[i]] = 1
			if i-l+1 > max {
				fmt.Printf("%d  %d\n", l, i)
				max = i - l + 1
			}
			continue
		}
		for l < i && m[s[i]] != 0 {
			delete(m, m[s[l]])
			l++
		}
		i--
	}
	return max
}
