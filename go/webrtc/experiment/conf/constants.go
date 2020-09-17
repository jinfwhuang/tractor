package conf

import (
	"net"
)

const (
	MaxDatagramSize = 8192
	Port = 10000
)

//EventBusAddr = "239.20.20.21"
//EventBusPort = 10000

var Group net.IP = net.IPv4(239, 20, 20, 21)
var Addr *net.UDPAddr = &net.UDPAddr {
	IP: Group,
	Port: Port,
}

