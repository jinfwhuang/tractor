package main

import (
	"github.com/farm-ng/tractor/webrtc/experiment/conf"
	"log"
	"net"
	"os"
	"time"

	//eventbusmain "github.com/farm-ng/tractor/webrtc/cmd/eventbus"
)


func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	os.Setenv("FARM_NG_ROOT", "/Users/jin/code/tractor")
	main1()
}

func main1() {
	msg := "default msg"
	if len(os.Args) > 1 {
		msg = os.Args[1]
	}
	log.Println(msg)

	conn, err := net.DialUDP("udp4", nil, conf.Addr)
	if err != nil {
		log.Fatalln(err)
	}
	go func() {
		for {
			conn.Write([]byte(msg))
			time.Sleep(1 * time.Second)
		}
	}()

	log.Println("waiting forever")
	select {}
}

