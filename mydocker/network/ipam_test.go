package network

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"testing"
)

func setIPAM_Load() {
	m := make(map[string][]uint64, 1)
	m["key1"] = []uint64{1, 2, 3}
	ipam := IPAM{
		IPAMConfigPath: "./ipam_test.json",
		Subnets:        &m,
	}
	if err := ipam.dump(); err != nil {
		panic(err)
	}
	if err := ipam.load(); err != nil {
		panic(err)
	}
	fmt.Println(ipam.Subnets)
}

func TestIPAM_LoadAndDump(t *testing.T) {
	setIPAM_Load()
}

func TestAllocate(t *testing.T) {
	m := make(map[string][]uint64, 1)
	ipam := IPAM{
		IPAMConfigPath: "./ipam_test.json",
		Subnets:        &m,
	}

	for true {
		ip, err := ipam.Allocate("192.168.1.0/24")
		if err != nil {
			fmt.Println(err)
			return
		} else {
			fmt.Println(ip)
		}
	}
}

func TestIPAMRelease(t *testing.T) {
	m := make(map[string][]uint64, 1)
	ipam := IPAM{
		IPAMConfigPath: "./ipam_test.json",
		Subnets:        &m,
	}
	if err := ipam.Release("192.168.1.0/24", "192.168.1.255"); err != nil {
		fmt.Println(err)
	}

}

func a() func() {
	fmt.Println("A")
	return func() {
		fmt.Println("B")
	}
}

func TestDefer(t *testing.T) {
	filepath.Walk("./", func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		_, file := filepath.Split(path)
		fmt.Println(file)
		return nil
	})
}

/*
172       .   16   .    0         .   0
10101100     00010000   00000000      00000000
00000000     00000001   00000000      00010011

*/
