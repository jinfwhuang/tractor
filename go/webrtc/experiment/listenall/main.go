package main

import (
	"encoding/hex"
	"github.com/farm-ng/tractor/webrtc/experiment/conf"
	"log"
	"net"
	"os"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	os.Setenv("FARM_NG_ROOT", "/Users/jin/code/tractor")

	main2()
}

func main2() {
	// Open a connection and join group
	lnConn, err := net.ListenMulticastUDP("udp4", nil, &net.UDPAddr {
		IP: net.IPv4(239, 20, 20, 20),
		Port: conf.Port,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Loop forever reading from the socket
	for {
		buffer := make([]byte, conf.MaxDatagramSize)
		numBytes, src, err := lnConn.ReadFromUDP(buffer)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}
		log.Println(numBytes, "bytes read from", src, hex.Dump(buffer[:numBytes]))
	}

}


