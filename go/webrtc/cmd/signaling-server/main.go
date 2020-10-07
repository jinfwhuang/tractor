package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/farm-ng/tractor/webrtc/internal/server"
	"github.com/gorilla/mux"
)

var farmNgRoot string
var serverAddr string
var serverPort string
var proxyEndpoint string

func init() {
	// TODO: remove @jin
	os.Setenv("FARM_NG_ROOT", "/Users/jin/code/tractor")


	farmNgRoot = os.Getenv("FARM_NG_ROOT")
	if farmNgRoot == "" {
		log.Fatalln("FARM_NG_ROOT must be set.")
	}

	// -port=8586
	flag.StringVar(&serverPort, "port", "8586", "")
	serverAddr = ":" + serverPort

	// -proxy=http://nanohost:8585
	flag.StringVar(&proxyEndpoint, "proxy", "http://nanohost:8585", "")
	serverAddr = ":" + serverPort
}

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	srv := &http.Server{
		Handler:      createRouter(),
		Addr:         serverAddr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Serving frontend and API at:", serverAddr)
	log.Fatal(srv.ListenAndServe())
}

func createRouter() *mux.Router {
	router := mux.NewRouter()

	//// signaling
	//signaling := server.CreateSignalingHandler(proxyEndpoint)
	//router.PathPrefix("/twirp/").Handler(*signaling)

	// signaling2
	signaling2 := server.CreateSignalingHandler2()
	router.PathPrefix("/twirp/").Handler(*signaling2)

	// spa
	staticPath := path.Join(farmNgRoot, "app/frontend/dist")
	spa := server.CreateSpaHandler(staticPath)
	router.PathPrefix("/").Handler(*spa)

	return router
}
