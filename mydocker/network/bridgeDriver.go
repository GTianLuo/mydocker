package network

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"my_docker/mydocker/common"
	"my_docker/mydocker/container"
	"net"
	"os"
	"runtime"
	"strings"
)

type BridgeNetworkDriver struct {
	NetworkDriver
}

// Create 创建bridge网桥
func (d *BridgeNetworkDriver) Create(subnet, ip, name string) (*NetWork, error) {
	// 192.161.1.1/24  ---> 192.161.1.1 和 192.161.1.0/24
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, fmt.Errorf("parse net error:%v", err)
	}
	ipNet.IP = net.ParseIP(ip)
	n := &NetWork{Name: name, IpRange: ipNet, Driver: "bridge"}
	// 初始化网桥
	err = d.initBridge(n)
	return n, err
}

func (d *BridgeNetworkDriver) Delete(name string) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("Get interface device %s error :%v", name, err)
	}
	return netlink.LinkDel(link)
}

// Connect 创建Veth(endpoint)，并将master一端到挂载到该网络
func (d *BridgeNetworkDriver) Connect(network *NetWork, endpoint *Endpoint) error {
	// 获取网桥信息
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}
	// 创建Veth接口的配置
	la := netlink.NewLinkAttrs()
	la.Name = endpoint.ID[:5]
	// Veth的master端挂载到bridge上
	la.Index = br.Attrs().Index

	//创建Veth对象
	endpoint.Device = netlink.Veth{
		LinkAttrs: la,
		PeerName:  "cif-" + la.Name,
	}
	if err := netlink.LinkAdd(&endpoint.Device); err != nil {
		return fmt.Errorf("Error Add Endpoint Device %v", err)
	}
	// ip link set up
	if err := netlink.LinkSetUp(&endpoint.Device); err != nil {
		return fmt.Errorf("Error set endpoint up: %v", err)
	}
	return nil
}

func (d *BridgeNetworkDriver) initBridge(n *NetWork) error {
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

// 进入容器的网络空间，并返回一个退出该netns的函数
func enterContainerNetns(link netlink.Link, cinfo *container.ContainerInfo) func() {
	file, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", cinfo.Pid), os.O_RDONLY, 0)
	if err != nil {
		log.Errorf("error get container net namespace,%v", err)
	}
	// 获取net文件的文件描述符
	fd := file.Fd()
	// 锁定线程，保证当前goroutine只在该线程上运行，并且当前线程只会运行该goroutine
	runtime.LockOSThread()
	// 将peerVeth 添加到指定netns
	if err := netlink.LinkSetNsFd(link, int(fd)); err != nil {
		log.Errorf("error set link netns, %v", err)
	}
	// 获取当前的ns，以便后面回到原本的ns
	origns, err := netns.Get()
	if err != nil {
		log.Errorf("error get current netns, %v", err)
	}
	if err := netns.Set(netns.NsHandle(fd)); err != nil {
		log.Errorf("error set netns, %v", err)
	}
	// 返回回到原先ns的函数
	return func() {
		netns.Set(origns)
		origns.Close()
		runtime.UnlockOSThread()
		file.Close()
	}
}

//配置容器的端口映射
func configPortMapping(ep *Endpoint, cinfo *container.ContainerInfo) error {
	for _, pm := range ep.PortMapping {
		portMapping := strings.Split(pm, ",")
		if len(portMapping) != 2 {
			//TODO 参数格式
			return fmt.Errorf("")
		}
		cmdArgs := fmt.Sprintf("-t nat -A PREROUTING -i %s -p tcp -m --dport %s -j DNAT --to-destination %s:%s",
			ep.Network.Name, portMapping[0], ep.IPAddress.String(), portMapping[1])
		errBytes, err := common.Exec("iptables", strings.Split(cmdArgs, " ")...)
		if err != nil {
			return fmt.Errorf("port mapping error:%s", string(errBytes))
		}
	}
	return nil
}

// 配置容器内的网络：将Veth挂载到该网络，并分配ip地址。配置默认路由信息
func configEndpointIpAddressAndRoute(ep *Endpoint, cinfo *container.ContainerInfo) error {
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return fmt.Errorf("configEndpointIpAddressAndRoute get link error:%v", err)
	}
	defer enterContainerNetns(peerLink, cinfo)()

	// 配置该端点的IP
	peerLinkIpnet := ep.Network.IpRange
	peerLinkIpnet.IP = ep.IPAddress

	if err := setInterfaceIp(ep.Device.PeerName, peerLinkIpnet.String()); err != nil {
		return fmt.Errorf("configEndpointIpAddressAndRoute set veth ip error:%v", err)
	}

	if err := setInterfaceUp(ep.Device.PeerName); err != nil {
		return fmt.Errorf("configEndpointIpAddressAndRoute set veth up error:%v", err)
	}
	if err := setInterfaceUp("lo"); err != nil {
		return fmt.Errorf("configEndpointIpAddressAndRoute set lo up error:%v", err)
	}

	// 配置默认路由
	// route add -net 0.0.0.0/0 gw {Bridge地址} dev {容器内的Veth端点地址}
	// 目标地址属于0.0.0.0/0网段时，请求交给网络设备Veth转发给Bridge
	// 0.0.0.0/0 表示所有的ip地址
	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        ep.Network.IpRange.IP,
		Dst:       cidr,
	}
	if err := netlink.RouteAdd(defaultRoute); err != nil {
		return fmt.Errorf("configEndpointIpAddressAndRoute add default route error:%v", err)
	}
	return nil
}
