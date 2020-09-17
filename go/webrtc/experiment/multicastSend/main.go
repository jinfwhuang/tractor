package main

import (
	"github.com/farm-ng/tractor/webrtc/experiment/conf"
	"golang.org/x/net/ipv4"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	os.Setenv("FARM_NG_ROOT", "/Users/jin/code/tractor")
	main2()
}

func main2() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	os.Setenv("FARM_NG_ROOT", "/Users/jin/code/tractor")

	msg := "default msg"
	if len(os.Args) > 1 {
		msg = os.Args[1]
	}
	log.Println(msg)

	ifi, err := net.InterfaceByName("en0")
	lo0, err := net.InterfaceByName("lo0")
	en0, err := net.InterfaceByName("en0")
	//ifi, err := net.InterfaceByName("en0")
	_ = lo0
	_ = en0
	_ = ifi

	// Opening connection
	var c net.PacketConn
	c, err = net.ListenPacket("udp4", "0.0.0.0:53333")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	var p *ipv4.PacketConn = ipv4.NewPacketConn(c)  // A richer version of net.PacketConn
	p.SetControlMessage(ipv4.FlagTTL, true)
	p.SetControlMessage(ipv4.FlagSrc, false)
	p.SetControlMessage(ipv4.FlagDst, true)
	p.SetControlMessage(ipv4.FlagDst, true)


	//c.WriteTo()

	// Joining group
	//if err := p.JoinGroup(lo0, &net.UDPAddr {
	//	IP: conf.Group,
	//	//Port: conf.Port,
	//}); err != nil {
	//	panic(err)
	//}
	//if err := p.JoinGroup(en0, &net.UDPAddr {
	//	IP: net.IPv4(239, 20, 20, 20),
	//	//Port: conf.Port,
	//}); err != nil {
	//	panic(err)
	//}
	//if err := p.JoinGroup(en0, &net.UDPAddr {
	//	IP: net.IPv4(239, 20, 20, 21),
	//	//Port: conf.Port,
	//}); err != nil {
	//	panic(err)
	//}
	//if err := p.SetControlMessage(ipv4.FlagSrc, true); err != nil {
	//	panic(err)
	//}
	//p.SetControlMessage(ipv4.FlagTTL, true)
	//p.SetControlMessage(ipv4.FlagSrc, false)
	//p.SetControlMessage(ipv4.FlagDst, true)
	//p.SetControlMessage(ipv4.FlagDst, true)
	//if err := p.SetControlMessage(ipv4.FlagDst, true); err != nil {
	//	panic(err)
	//}
	//log.Println(p)


	// Loop forever to publish data
	go func() {
		for {
			cm := ipv4.ControlMessage {
				//Src: net.IPv4(0, 0, 0, 10),
					//Dst: net.IPv4(100, 100, 100, 9),
				//Dst: conf.Group,
			}
			//if _, err := p.WriteTo([]byte(msg), &cm, &net.UDPAddr {
			//	IP: conf.Group,
			//	Port: conf.Port,
			//}); err != nil {
			//	log.Fatal(err)
			//}

			if _, err := p.WriteTo([]byte(msg), &cm, &net.UDPAddr {
				IP: net.IPv4(239, 20, 20, 20),
				Port: conf.Port,
			}); err != nil {
				log.Fatal(err)
			}

			log.Println(cm)
			log.Println("send msg", msg)
			time.Sleep(1 * time.Millisecond)
		}
	}()


	//// Loop forever reading from the socket
	//go func() {
	//	for {
	//		buffer := make([]byte, 1500)
	//		log.Println("starting to read from", p)
	//		n, cm, src, err := p.ReadFrom(buffer)
	//		if err != nil {
	//			log.Fatal("ReadFromUDP failed:", err)
	//		}
	//		log.Println(cm)
	//		log.Println(n, "bytes read from", src, hex.Dump(buffer[:n]))
	//	}
	//}()


	log.Println("wiat for ever")
	select {}
}

