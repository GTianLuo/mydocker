package network

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"my_docker/mydocker/common"
	"net"
	"os"
	"path"
)

const ipamConfigDefaultPath = "/var/lib/mydocker/network/subnet.json"

type IPAM struct {
	IPAMConfigPath string
	Subnets        *map[string][]uint64
}

var DefaultIPAM = IPAM{
	IPAMConfigPath: ipamConfigDefaultPath,
}

// load 加载IPAM配置文件中的内容
func (ipam *IPAM) load() error {
	// 判断配置文件是否存在
	if exist, err := common.PathExist(ipam.IPAMConfigPath); err != nil {
		return fmt.Errorf("ipam load error:%v", err)
	} else if !exist {
		// 不存在
		return nil
	}
	// 打开文件，并读取配置文件的内容
	file, err := os.Open(ipam.IPAMConfigPath)
	if err != nil {
		return fmt.Errorf("ipam load open file error:%v", err)
	}
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("ipam load read file error:%v", err)
	}
	if err := json.Unmarshal(bytes, ipam.Subnets); err != nil {
		return fmt.Errorf("ipam load unmarshal error:%v", err)
	}
	return nil
}

// dump 持久化存储IPAM
func (ipam *IPAM) dump() error {
	dir, _ := path.Split(ipam.IPAMConfigPath)
	if err := common.MkdirIfNotExist(dir); err != nil {
		return fmt.Errorf("ipam dump mkdir error:%v", err)
	}
	file, err := os.OpenFile(ipam.IPAMConfigPath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("ipam dump open file error:%v", err)
	}
	jsonBytes, err := json.Marshal(ipam.Subnets)
	if err != nil {
		return fmt.Errorf("ipam dupm marshal error:%v", err)
	}
	_, err = file.Write(jsonBytes)
	if err != nil {
		return fmt.Errorf("ipam dump write file error:%v", err)
	}
	return nil
}

func (ipam *IPAM) Allocate(subnetStr string) (ipStr string, err error) {
	_, subnet, err := net.ParseCIDR(subnetStr)
	if err != nil {
		return "", fmt.Errorf("ipam allocate error:%v", err)
	}
	ipam.Subnets = &map[string][]uint64{}
	err = ipam.load()
	if err != nil {
		return "", fmt.Errorf("ipam allocate error:%v", err)
	}
	// ones 是掩码占的位数，bits是总位数
	ones, bits := subnet.Mask.Size()
	// counts是该ip段最多分配的ip数量
	var counts uint32 = 1<<uint8(bits-ones) - 1
	if _, exist := (*ipam.Subnets)[subnet.String()]; !exist {
		// 不存在，说明该网段是首次分配
		// counts 该网段能表示的ip总数
		(*ipam.Subnets)[subnet.String()] = getBitmap(counts)
	}
	bitmap := (*ipam.Subnets)[subnet.String()]
	ip := subnet.IP
	for i, m := range bitmap {
		var c uint64 = 1
		ii := 1
		for ii <= 64 && uint32(64*i+ii) <= counts {
			if (m & c) == 0 {
				//找到了未使用的ip
				bitmap[i] = m | c
				n := 64*i + ii
				for i := 3; i >= 0; i-- {
					[]byte(ip)[i] += uint8(n >> ((3 - i) * 8))
				}
				return ip.String(), ipam.dump()
			}
			ii++
			c = c << 1
		}
	}
	return "", fmt.Errorf("ipam allocate error: IP has been used up")

}

// Release 释放地址
func (ipam *IPAM) Release(subnetStr string, ipStr string) error {
	ip := net.ParseIP(ipStr)
	_, subnet, err := net.ParseCIDR(subnetStr)
	if err != nil {
		return fmt.Errorf("ipam release error:%v", err)
	}
	// ip是否属于这个网段
	if !subnet.Contains(ip) {
		return fmt.Errorf("ipam release error:%s does not belong to the network segment %s", ipStr, subnetStr)
	}
	if err := ipam.load(); err != nil {
		return fmt.Errorf("ipam release error:%v", err)
	}
	if len((*ipam.Subnets)[subnetStr]) == 0 {
		return nil
	}
	// 计算该ip在网段中的位置
	var count uint32 = 0
	for i := 3; i >= 0; i-- {
		count += uint32(ip.To4()[i]-subnet.IP.To4()[i]) << ((3 - i) * 8)
	}
	bitmap := (*ipam.Subnets)[subnetStr]
	i := (count / 64) - 1
	if count%64 != 0 {
		i++
	}
	bitmap[i] &= ^(1 << ((count - 1) % 64))
	if err := ipam.dump(); err != nil {
		return fmt.Errorf("ipam release dump error:%v ", err)
	}
	return nil
}

// 通过bitNum获取大小合适的位图
func getBitmap(bitNum uint32) []uint64 {
	count := bitNum / 64
	if bitNum%64 != 0 {
		count++
	}
	return make([]uint64, count)
}
