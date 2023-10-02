package network

type NetworkDriver interface {
	Name() string
	Create(subnet, ip, name string) (*NetWork, error)
	Delete(name string) error

	// Connect 连接一个网络和网络端点
	Connect(network *NetWork, endpoint *Endpoint) error
}
