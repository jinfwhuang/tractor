package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/rs/cors"

	"github.com/farm-ng/tractor/genproto"
	pb "github.com/farm-ng/tractor/genproto"
	"github.com/farm-ng/tractor/webrtc/internal/api"
	"github.com/farm-ng/tractor/webrtc/internal/eventbus"
	"github.com/farm-ng/tractor/webrtc/internal/proxy"
	"github.com/farm-ng/tractor/webrtc/internal/spa"
	"github.com/gorilla/mux"

	 //eventbusmain "github.com/farm-ng/tractor/webrtc/cmd/eventbus"
)

const (
	EventBusAddr = "239.20.20.21"
	EventBusPort = 10000
	RtpAddr      = "239.20.20.20:5000"
	// Set this too low and we see packet loss in chrome://webrtc-internals, and on the network interface (`netstat -suna`)
	// But what should it be? `sysctl net.core.rmem_max`?
	rtpReadBufferSize  = 1024 * 1024 * 8
	maxRtpDatagramSize = 4096
	defaultServerAddr  = ":8080"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	os.Setenv("FARM_NG_ROOT", "/Users/jin/code/tractor")

	// Create EventBus proxy
	eventChan := make(chan *pb.Event)
	eventBus := eventbus.NewEventBus((net.UDPAddr{IP: net.ParseIP(EventBusAddr), Port: EventBusPort}), "webrtc-proxy", eventChan, true)
	eventBusProxy := proxy.NewEventBusProxy((&proxy.EventBusProxyConfig{EventBus: eventBus, EventSource: eventChan}))

	// Create RTP proxy
	rtpProxy := proxy.NewRtpProxy(&proxy.RtpProxyConfig{ListenAddr: RtpAddr, ReadBufferSize: rtpReadBufferSize, MaxDatagramSize: maxRtpDatagramSize})

	// Create and start webRTC proxy
	proxy := proxy.NewProxy(eventBusProxy, rtpProxy)
	//proxy.Start()

	// Create API handler
	server := api.NewServer(proxy)
	twirpHandler := genproto.NewWebRTCProxyServiceServer(server, nil)
	corsWrapper := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"POST"},
		AllowedHeaders: []string{"Content-Type"},
	})
	api := corsWrapper.Handler(twirpHandler)

	// Create SPA handler
	farmNgRoot := os.Getenv("FARM_NG_ROOT")
	if farmNgRoot == "" {
		log.Fatalln("FARM_NG_ROOT must be set.")
	}
	spa := spa.Handler{StaticPath: path.Join(farmNgRoot, "build/frontend"), IndexPath: "index.html"}

	// Serve the API and frontend
	serverAddr := defaultServerAddr
	port := os.Getenv("PORT")
	if port != "" {
		serverAddr = ":" + port
	}
	router := mux.NewRouter()
	router.PathPrefix("/twirp/").Handler(api)
	router.PathPrefix("/").Handler(spa)
	srv := &http.Server{
		Handler:      router,
		Addr:         serverAddr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Serving frontend and API at:", serverAddr)
	log.Fatal(srv.ListenAndServe())
}
