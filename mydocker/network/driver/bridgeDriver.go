package main

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"my_docker/mydocker/common"
	"my_docker/mydocker/network"
	"net"
	"strings"
)

type BridgeNetworkDriver struct {
	NetworkDriver
}

// Create 创建bridge网桥
func (d *BridgeNetworkDriver) Create(subnet string, name string) (*network.NetWork, error) {
	// 192.161.1.1/24  ---> 192.161.1.1 和 192.161.1.0/24
	ip, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, fmt.Errorf("parse net error:%v", err)
	}
	ipNet.IP = ip
	n := &network.NetWork{Name: name, IpRange: ipNet, Driver: "bridge"}
	// 初始化网桥
	err = d.initBridge(n)
	return n, err
}

func (d *BridgeNetworkDriver) initBridge(n *network.NetWork) error {
	if err := createBridgeInterface(n.Name); err != nil {
		return err
	}
	if err := setInterfaceIp(n.Name, n.IpRange.String()); err != nil {
		return err
	}
	if err := setInterfaceUp(n.Name); err != nil {
		return err
	}
	return setUpIptables(n.Name, n.IpRange)
}

// 创建Bridge 虚拟网络设备
func createBridgeInterface(bridgeName string) error {
	// 判断bridge是否存在
	if exist, err := common.DevInterIsExist(bridgeName); !exist {
		if err == nil {
			err = fmt.Errorf("bridge %s has exist", bridgeName)
		}
		return fmt.Errorf("create bridge interface error: %v", err)
	}
	// 创建一个默认的link对象
	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = bridgeName
	// 创建bridge对新
	br := &netlink.Bridge{LinkAttrs: linkAttrs}
	if err := netlink.LinkAdd(br); err != nil {
		return fmt.Errorf("Bridge %s creation error:%v", bridgeName, err.Error())
	}
	return nil
}

// 设置网络接口设备的ip
// 给bridge分配一个ip后，会自动配置路由表。
// 例：配置ip: 10.1.1.1/24
// 会自动配置路由表，网段10.1.1.0/24都会转发到bridge
func setInterfaceIp(name string, rawIP string) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("Get interface device %s error:%v", name, err)
	}
	// 将“192.168.0.1/24”这样的字符串解析成ip和mask
	ipNet, err := netlink.ParseIPNet(rawIP)
	if err != nil {
		return err
	}

	return netlink.AddrAdd(link, &netlink.Addr{IPNet: ipNet, Label: "", Flags: 0, Scope: 0})
}

// ip link XXX up
func setInterfaceUp(name string) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("Get interface device %s error :%v", name, err)
	}
	if err = netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("Set interface device %s error :%v", name, err)
	}
	return nil
}

// 设置iptables 对应bridge的MASQUERADE规则
func setUpIptables(bridgeName string, subnet *net.IPNet) error {
	// iptables -t nat -A POSTROUTING -s <IPNet> ! -o <bridgeName> -j MASQUERADE
	// -t nat指定命令应用于内置表nat
	// -A POSTROUTING 在POSTROUTING链上追加命令
	// -s 指定应用该规则的数据包的源地址
	// ! 表示反
	// -o 指定应用该规则的数据包的出口网卡
	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(), bridgeName)
	if errBytes, err := common.Exec("iptables", strings.Split(iptablesCmd, " ")...); err != nil {
		return fmt.Errorf("set iptables error:%s", string(errBytes))
	}
	return nil
}

func main() {
	d := &BridgeNetworkDriver{}
	_, err := d.Create("10.1.1.1/24", "br0")
	fmt.Println(err)
}
