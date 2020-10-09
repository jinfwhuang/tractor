package server

import (
	"github.com/farm-ng/tractor/genproto"
	"github.com/farm-ng/tractor/webrtc/internal/api"
	"github.com/farm-ng/tractor/webrtc/internal/proxy"
	"github.com/farm-ng/tractor/webrtc/internal/spa"
	"github.com/rs/cors"
	"net/http"
)

func CreateSpaHandler(staticPath string) *spa.Handler {
	//spa := spa.Handler{StaticPath: path.Join(farmNgRoot, "build/frontend"), IndexPath: "index.html"}
	spa := spa.Handler{
		StaticPath: staticPath,
		IndexPath: "index.html",
	}
	return &spa
}

func CreateBlobstoreHandler(path string) *http.Handler {
	blobstoreCorsWrapper := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{"Content-Type"},
	})
	blobstore := blobstoreCorsWrapper.Handler(
		http.FileServer(http.Dir(path)))
	return &blobstore
}

func CreateApiHandler(proxy *proxy.Proxy) *http.Handler {
	server := api.NewServer(proxy)
	twirpHandler := genproto.NewWebrtcApiServiceServer(server, nil)
	corsWrapper := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"POST"},
		AllowedHeaders: []string{"Content-Type"},
	})
	api := corsWrapper.Handler(twirpHandler)
	return &api
}

func CreateSignalingHandler() *http.Handler {
	server := &SignalingServer{
		//proxyEndpoint: proxyEndpoint,
	}
	twirpHandler := genproto.NewWebrtcApiServiceServer(server, nil)
	corsWrapper := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"POST"},
		AllowedHeaders: []string{"Content-Type"},
	})
	api := corsWrapper.Handler(twirpHandler)
	return &api
}

