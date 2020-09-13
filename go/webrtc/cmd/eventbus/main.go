package main

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"time"

	"github.com/farm-ng/tractor/webrtc/internal/eventbus"
	proxyServer "github.com/farm-ng/tractor/webrtc/cmd/proxy-server"

)

const (
	FFF string = "239.20.20.21"
)

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "  ")
	return string(s)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	os.Setenv("FOO", "1")

	//udpAddress := "0.0.0.0"
	//udpAddress = "239.20.20.21"
	//eventbus.
	udpAddress := proxyServer.EventBusAddr
	port := proxyServer.EventBusPort

	b := eventbus.NewEventBus(net.UDPAddr{IP: net.ParseIP(udpAddress), Port: port}, "go-eventbus", nil, false)

	stateTicker := time.NewTicker(1 * time.Second)
	announcementsTicker := time.NewTicker(10 * time.Second)

	go func() {
		for {
			select {
			case <-stateTicker.C:
				log.Println("State", prettyPrint(b.State))
			case <-announcementsTicker.C:
				log.Println("Announcements", prettyPrint(b.Announcements))
			}
		}
	}()
	b.Start()
}
