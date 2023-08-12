package main

import "my_docker/mydocker/network"

type NetworkDriver interface {
	Name() string
	Create(subnet string, name string) (*network.NetWork, error)
}
