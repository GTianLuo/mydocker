package network

import (
	"github.com/vishvananda/netlink"
	"net"
)

// NetWork 网络，多个容器可以共享一个网络
type NetWork struct {
	Name    string     // 网络名
	IpRange *net.IPNet // 表示一个网段
	Driver  string     // 驱动名
}

// Endpoint 网络端点 用于容器和网络的连接
type Endpoint struct {
	ID          string           `json:"id"`
	Device      netlink.Veth     `json:"dev"` // 容器的虚拟网卡设备
	IPAddress   net.IP           `json:"ip"`  //
	MacAddress  net.HardwareAddr `json:"mac"`
	PortMapping []string         `json:"portmapping"`
	Network     *NetWork         `json:"network"`
}
