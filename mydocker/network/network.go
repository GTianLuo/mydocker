package network

import (
	"encoding/json"
	"fmt"
	"github.com/vishvananda/netlink"
	"io/fs"
	"io/ioutil"
	"my_docker/mydocker/common"
	"my_docker/mydocker/container"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const networkConfigPath = "/var/lib/mydocker/network/"

const networkConfigFile = "%s.json"

var drivers = map[string]NetworkDriver{}

var networks = map[string]*NetWork{}

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

// Init 初始化驱动和网络配置
func Init() error {
	bridgeDriver := &BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = bridgeDriver

	if err := common.MkdirIfNotExist(networkConfigPath); err != nil {
		return fmt.Errorf("network init error:%v", err)
	}
	return filepath.Walk(networkConfigPath, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		//
		_, fileName := filepath.Split(path)
		network := &NetWork{
			Name: strings.Split(fileName, ".")[0],
		}
		if err := network.load(); err != nil {
			return err
		}
		networks[network.Name] = network
		return nil
	})
}

func (nw *NetWork) dump() error {
	if err := common.MkdirIfNotExist(networkConfigPath); err != nil {
		return fmt.Errorf("network dump error:%v", err)
	}
	dumpFile, err := os.OpenFile(networkConfigPath+fmt.Sprintf(networkConfigFile, nw.Name),
		os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0777)
	defer dumpFile.Close()
	if err != nil {
		return fmt.Errorf("network dump error:%v", err)
	}
	jsonBytes, err := json.Marshal(nw)
	if err != nil {
		return fmt.Errorf("network dump marshal error:%v", err)
	}
	_, err = dumpFile.Write(jsonBytes)
	if err != nil {
		return fmt.Errorf("network dump write file error:%v", err)
	}
	return nil
}

func (nw *NetWork) load() error {
	jsonByte, err := ioutil.ReadFile(networkConfigPath + fmt.Sprintf(networkConfigFile, nw.Name))
	if err != nil {
		return fmt.Errorf("network load read file error:%v", err)
	}
	if err = json.Unmarshal(jsonByte, nw); err != nil {
		return fmt.Errorf("network load unmarshal error:%v", err)
	}
	return nil
}

func CreateNetwork(driver, subnet, name string) error {
	gatewayIp, err := DefaultIPAM.Allocate(subnet)
	if err != nil {
		return err
	}
	nw, err := drivers[driver].Create(subnet, gatewayIp, name)
	if err != nil {
		return err
	}
	return nw.dump()
}

func Connect(networkName string, cinfo *container.ContainerInfo) error {
	network, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("No such NetWork:%s", networkName)
	}
	// 通过IPAM从网络中分配可用ip
	ipStr, err := DefaultIPAM.Allocate(network.IpRange.String())
	if err != nil {
		return err
	}
	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", cinfo.Id, networkName),
		IPAddress:   net.ParseIP(ipStr),
		PortMapping: cinfo.PortMapping,
		Network:     network,
	}
	// 调用网络驱动的connect方法配置网络和端点,会为端点创建一对Veth，并将一端挂载到bridge上
	if err := drivers[network.Driver].Connect(network, ep); err != nil {
		return err
	}
	// 进入容器netns，配置容器的Veth，ip和路由,函数在结束时，会退出容器netns
	if err := configEndpointIpAddressAndRoute(ep, cinfo); err != nil {
		return err
	}
	// 使用DNAT配置容器到宿主机的端口映射
	return configPortMapping(ep, cinfo)
}
