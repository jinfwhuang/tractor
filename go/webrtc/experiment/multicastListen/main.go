package main

import (
	"encoding/hex"
	"fmt"

	//"fmt"
	"github.com/farm-ng/tractor/webrtc/experiment/conf"
	"golang.org/x/net/ipv4"
	"golang.org/x/sys/unix"
	"log"
	"net"
	"os"
	"syscall"
	"context"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	os.Setenv("FARM_NG_ROOT", "/Users/jin/code/tractor")

	main2()
}

func main2() {
	ifi, _ := net.InterfaceByName("en0")
	lo0, _ := net.InterfaceByName("lo0")
	en0, _ := net.InterfaceByName("en0")
	//ifi, err := net.InterfaceByName("en0")
	_ = lo0
	_ = en0
	_ = ifi

	listenConfig := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var err error
			c.Control(func(fd uintptr) {
				err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
				if err != nil {
					return
				}

				err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
				if err != nil {
					return
				}
			})
			return err
		},
	}

	c, err := listenConfig.ListenPacket(context.Background(), "udp4", fmt.Sprintf(":%d", 10000))

	log.Println(err)
	pp := ipv4.NewPacketConn(c)

	//ipv4.PacketConn


	//c.ReadFrom()

	// Joining group
	p := ipv4.NewPacketConn(c)
	//if err := p.JoinGroup(lo0, &net.UDPAddr {
	//	IP: conf.Group,
	//	//Port: conf.Port,
	//}); err != nil {
	//	panic(err)
	//}
	//if err := p.JoinGroup(lo0, &net.UDPAddr {
	//	IP: net.IPv4(239, 20, 20, 20),
	//	//Port: 10000,
	//}); err != nil {
	//	panic(err)
	//}
	if err := p.JoinGroup(en0, &net.UDPAddr {
		IP: net.IPv4(239, 20, 20, 20),
		//Port: 10000,
	}); err != nil {
		panic(err)
	}
	//if err := p.JoinGroup(en0, &net.UDPAddr {
	//	IP: conf.Group,
	//	//Port: conf.Port,
	//}); err != nil {
	//	panic(err)
	//}

	p.SetControlMessage(ipv4.FlagTTL, true)
	p.SetControlMessage(ipv4.FlagSrc, false)
	p.SetControlMessage(ipv4.FlagDst, true)

	pp.SetControlMessage(ipv4.FlagTTL, true)
	pp.SetControlMessage(ipv4.FlagSrc, false)
	pp.SetControlMessage(ipv4.FlagDst, true)
	//if err := pp.SetControlMessage(ipv4.FlagSrc, true); err != nil {
	//	panic(err)
	//}
	//if err := p.SetControlMessage(ipv4.FlagDst, true); err != nil {
	//	panic(err)
	//}
	log.Println(p)

	// Loop forever reading from the socket
	for {
		buffer := make([]byte, conf.MaxDatagramSize)
		numBytes, cm, src, err := p.ReadFrom(buffer)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}
		log.Println(cm)
		//log.Println(cm.Dst.IsMulticast())
		//log.Println(cm.Dst.IsMulticast())
		log.Println(numBytes, "bytes read from", src, hex.Dump(buffer[:numBytes]))
	}

}


