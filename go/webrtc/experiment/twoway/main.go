package main

import (
	"encoding/hex"
	"fmt"
	"golang.org/x/net/ipv4"
	"log"
	"net"
	"os"
	"time"

	//eventbusmain "github.com/farm-ng/tractor/webrtc/cmd/eventbus"
)

const (
	address = "239.20.20.20:10000"
	maxDatagramSize = 8192

	//EventBusAddr = "239.20.20.21"
	//EventBusPort = 10000
	EventBusAddr = "239.0.0.0"
	EventBusPort = 9999

	RtpAddr      = "239.20.20.20:5000"
	// Set this too low and we see packet loss in chrome://webrtc-internals, and on the network interface (`netstat -suna`)
	// But what should it be? `sysctl net.core.rmem_max`?
	rtpReadBufferSize  = 1024 * 1024 * 8
	maxRtpDatagramSize = 4096
	defaultServerAddr  = ":8080"
)

func main() {
	main2()
}

func main2() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	os.Setenv("FARM_NG_ROOT", "/Users/jin/code/tractor")

	//ip := net.IPv4(230, 20, 20, 20)

	msg := "default msg"
	if len(os.Args) > 1 {
		msg = os.Args[1]
	}
	log.Println(msg)

	addr, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		log.Fatal(err)
	}

	// Open up a listening UPD connection
	conn, err := net.ListenMulticastUDP("udp4", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	conn.SetReadBuffer(maxDatagramSize)
	go func() {
		for {
			buffer := make([]byte, maxDatagramSize)
			numBytes, src, err := conn.ReadFromUDP(buffer)
			if err != nil {
				log.Fatal("ReadFromUDP failed:", err)
			}
			log.Println(numBytes, "bytes read from", src, hex.Dump(buffer[:numBytes]))
		}
	}()

	// Open up a broadcast UDP connection
	addr, err = net.ResolveUDPAddr("udp4", address)
	if err != nil {
		log.Fatal("cannot resolve address", err)
	}
	sendConn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		log.Fatal("cannot open a broadcast connection", err)
	}
	go func() {
		for {
			sendConn.Write([]byte(msg))
			time.Sleep(1 * time.Second)
		}
	}()


	log.Println("waiting forever")
	select {}
}

func main1() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	os.Setenv("FARM_NG_ROOT", "/Users/jin/code/tractor")

	msg := "default msg"
	if len(os.Args) > 1 {
		msg = os.Args[1]
	}
	//if msg == nil {
	//	msg = "default msg"
	//}
	log.Println(msg)

	sendConn, err := NewBroadcaster(address)
	if err != nil {
		log.Fatal(err)
	}
	recConn := ipv4.NewPacketConn(sendConn)

	//go func() {
	//	for {
	//		sendConn.Write([]byte(msg))
	//		time.Sleep(1 * time.Second)
	//	}
	//}()

	// Loop forever reading from the socket
	fmt.Println("receiving conn", recConn)

	go func() {
		for {
			fmt.Println(1)
			buffer := make([]byte, maxDatagramSize)
			//recConn.rea
			fmt.Println(2)
			numBytes, _, src, err := recConn.ReadFrom(buffer)
			if err != nil {
				log.Fatal("ReadFromUDP failed:", err)
			}
			fmt.Println(3)

			log.Println(numBytes, "bytes read from", src, hex.Dump(buffer[:numBytes]))
		}
	}()

	//sendConn
	//sendConn.JoinGroup(nil, &net.UDPAddr{IP: bus.multicastGroup.IP})

	////c, err := listenConfig.ListenPacket(context.Background(), "udp4", fmt.Sprintf(":%d", bus.multicastGroup.Port))
	////if err != nil {
	////	log.Fatalf("could not create receiveConn: %v", err)
	////}
	////defer c.Close()
	//
	//
	//// https://godoc.org/golang.org/x/net/ipv4#PacketConn.JoinGroup
	//// JoinGroup uses the system assigned multicast interface when ifi is nil,
	//// although this is not recommended...
	//p := ipv4.NewPacketConn(bus.receiveConn)
	//
	//err = p.JoinGroup(nil, &net.UDPAddr{IP: bus.multicastGroup.IP})
	//if err != nil {
	//	log.Printf("attemped to join udp multicast at: %v", bus.multicastGroup)
	//	log.Fatalf("receiveConn could not join group: %v", err)
	//}
	//
	//
	//
	log.Println("waiting forever")
	select {}
}


func NewBroadcaster(address string) (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		return nil, err
	}

	return conn, nil

}
