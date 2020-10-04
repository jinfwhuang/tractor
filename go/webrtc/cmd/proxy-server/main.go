package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"github.com/farm-ng/tractor/genproto"
	pb "github.com/farm-ng/tractor/genproto"
	"github.com/farm-ng/tractor/webrtc/cmd/conf"
	"github.com/farm-ng/tractor/webrtc/internal/api"
	"github.com/farm-ng/tractor/webrtc/internal/eventbus"
	"github.com/farm-ng/tractor/webrtc/internal/proxy"
	"github.com/farm-ng/tractor/webrtc/internal/spa"
)

const (
	rtpAddr = "127.0.0.1:5000"
	// Set this too low and we see packet loss in chrome://webrtc-internals, and on the network interface (`netstat -suna`)
	// But what should it be? `sysctl net.core.rmem_max`?
	rtpReadBufferSize  = 1024 * 1024 * 8
	maxRtpDatagramSize = 4096
	defaultServerAddr  = ":8585"
)

var flagApiServer bool
var flagBlobstore bool
var flagFrontendServer bool
var flagSignalServer bool
var farmNgRoot string
var serverAddr string

func init() {
	flag.BoolVar(&flagApiServer, "flagApiServer", true, "API Server")
	flag.BoolVar(&flagSignalServer, "flagSignalServer", true, "")
	flag.BoolVar(&flagFrontendServer, "flagFrontendServer", true, "Frontend Server")
	flag.BoolVar(&flagBlobstore, "flagBlobstore", true, "")
	flag.Parse()
	log.Println(flagApiServer)
	log.Println(flagSignalServer)
	log.Println(flagFrontendServer)

	farmNgRoot = os.Getenv("FARM_NG_ROOT")
	if farmNgRoot == "" {
		log.Fatalln("FARM_NG_ROOT must be set.")
	}

	serverAddr = defaultServerAddr
	port := os.Getenv("PORT")
	if port != "" {
		serverAddr = ":" + port
	}
}

func main() {
	proxy := startProxy()

	srv := &http.Server{
		Handler:      createRouter(proxy),
		Addr:         serverAddr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Serving frontend and API at:", serverAddr)
	log.Fatal(srv.ListenAndServe())
}


func createRouter(proxy *proxy.Proxy) *mux.Router {
	router := mux.NewRouter()
	if flagFrontendServer {
		spa := createSpaHandler()
		router.PathPrefix("/app").Handler(*spa)
	}
	if flagBlobstore {
		blobstore := createBlobstoreHandler()
		router.PathPrefix("/resources/").Handler(http.StripPrefix("/resources", *blobstore))
	}
	if flagApiServer {
		api := createApiHandler(proxy)
		router.PathPrefix("/twirp/").Handler(*api)
	}
	return router
}

func createSpaHandler() *spa.Handler {
	spa := spa.Handler{StaticPath: path.Join(farmNgRoot, "build/frontend"), IndexPath: "index.html"}
	return &spa
}

func createBlobstoreHandler() *http.Handler {
	blobstoreCorsWrapper := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{"Content-Type"},
	})
	blobstore := blobstoreCorsWrapper.Handler(
		http.FileServer(http.Dir(path.Join(farmNgRoot, "..", "tractor-data"))))
	return &blobstore
}

func createApiHandler(proxy *proxy.Proxy) *http.Handler {
	server := api.NewServer(proxy)
	twirpHandler := genproto.NewWebRTCProxyServiceServer(server, nil)
	corsWrapper := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"POST"},
		AllowedHeaders: []string{"Content-Type"},
	})
	api := corsWrapper.Handler(twirpHandler)
	return &api
}

func startProxy() *proxy.Proxy {
	// Create EventBus proxy
	eventChan := make(chan *pb.Event)
	eventBus := eventbus.NewEventBus(&eventbus.EventBusConfig{
		MulticastGroup: net.UDPAddr{
			IP:   net.ParseIP(conf.EventBusAddr),
			Port: conf.EventBusPort,
		},
		ServiceName: "webrtc-proxy",
	}).WithEventChannel(&eventbus.EventChannelConfig{
		Channel:              eventChan,
		PublishAnnouncements: true,
	})
	eventBusProxy := proxy.NewEventBusProxy(&proxy.EventBusProxyConfig{
		EventBus:    eventBus,
		EventSource: eventChan,
	})

	// Create Rtp proxy
	rtpProxy := proxy.NewRtpProxy(&proxy.RtpProxyConfig{
		ListenAddr:      rtpAddr,
		ReadBufferSize:  rtpReadBufferSize,
		MaxDatagramSize: maxRtpDatagramSize,
	})

	// Start webRTC proxy
	proxy := proxy.NewProxy(eventBusProxy, rtpProxy)
	proxy.Start()

	return proxy
}
