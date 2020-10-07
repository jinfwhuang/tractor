package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/farm-ng/tractor/webrtc/internal/common"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/protobuf/proto"

	"github.com/pion/datachannel"
	"github.com/pion/webrtc/v3"
	"github.com/twitchtv/twirp"
	"io"
	pb "github.com/farm-ng/tractor/genproto"
	//"github.com/pion/webrtc/v3/signal"
	//"github.com/pion/webrtc/v3/examples/internal/signal"
	"log"
)

//// Server is a Twirp server that exposes a webRTC signaling endpoint
//type SignalingServer struct {
//	//proxy *proxy.Proxy
//	proxyEndpoint string
//}

type Req struct {
	Sdp string `json:"sdp"`
}

//// TODO: It is hacky. It abuses the WebRTCProxyService definition. The convenience to reusing
////       the service is that the frontend app does not need to be configured whether it is
////       talking to API server or SIGNALING server
////
//// InitiatePeerConnection starts the proxy and returns an SDP answer to the client
//func (s *SignalingServer) InitiatePeerConnection(
//	ctx context.Context,
//	req *pb.InitiatePeerConnectionRequest) (res *pb.InitiatePeerConnectionResponse, err error) {
//
//	// Forward the request to the proxy-server
//	client := pb.NewWebRTCProxyServiceProtobufClient(s.proxyEndpoint, &http.Client{})
//	resp, err := client.InitiatePeerConnection(ctx, req)
//
//	log.Println("req", req.String())
//	log.Println("resp", resp.String())
//
//	return resp, nil
//}
//
//
//type ClientSignalingServer struct {
//	signalingEndpoint string
//}

const (
	signalingMessageSizeMax = 1000
)

// ------------------------------------------------
// Proxy -> Signaling
// ------------------------------------------------
// Server is a Twirp server that exposes a webRTC signaling endpoint

/*
Assumptions
1. Only talk to a single proxy server
2. Set up a WebRtc datachannel with a single proxy server
*/
type SignalingServer struct {
	//signalingEndpoint string
	/*
		1. holds the manager

	*/
	pc *webrtc.PeerConnection
	raw *datachannel.ReadWriteCloser
}

func NewSignalingServer() *SignalingServer {

	return nil
}

func (s *SignalingServer) emitSignalingEvent(connId string, req *pb.InitiatePeerConnectionRequest) {
	//message := req.String()

	event := &pb.WebRtcConnection{
		Stamp: ptypes.TimestampNow(),
		ConnId: connId,
		ClientSdp: req.Sdp,
		//ProxySdp:
	}
	eventBytes, err := proto.Marshal(event)

	d := *s.raw
	_, err = d.Write(eventBytes)

	log.Println("event size", len(eventBytes))

	if err != nil {
		panic(err)
	}
}

//// WriteLoop shows how to write to the datachannel directly
//func (s *SignalingServer) WriteLoop(d io.Writer) {
//	for range time.NewTicker(5 * time.Second).C {
//		message := strconv.Itoa(rand.Int())
//		log.Printf("Sending %s \n", message)
//
//		_, err := d.Write([]byte(message))
//		if err != nil {
//			panic(err)
//		}
//	}
//}

func (s *SignalingServer) processSignalingEvent() {

}

// TODO: It is hacky. It abuses the WebRTCProxyService definition. The convenience to reusing
//       the service is that the frontend app does not need to be configured whether it is
//       talking to API server or SIGNALING server
//
// InitiatePeerConnection starts the proxy and returns an SDP answer to the client
func (s *SignalingServer) InitiatePeerConnection(
	ctx context.Context,
	req *pb.InitiatePeerConnectionRequest,
) (res *pb.InitiatePeerConnectionResponse, err error) {

	//log.Fatal("not implemented")
	//return nil, nil

	peerId := common.UniquePeerID()
	s.emitSignalingEvent(peerId, req)

	// TODO: look for the response in some data structure

	// Reply
	//b, err := base64.StdEncoding.DecodeString(req.Sdp)
	//b, err = json.Marshal(answer)
	//if err != nil {
	//	return nil, twirp.NewError(twirp.Internal, "could not generate SDP")
	//}
	//Sdp := base64.StdEncoding.EncodeToString(b)
	return &pb.InitiatePeerConnectionResponse{
		Sdp: req.Sdp,
	}, nil

	//// Forward the request to the proxy-server
	//client := pb.NewWebRTCProxyServiceProtobufClient(s.proxyEndpoint, &http.Client{})
	//resp, err := client.InitiatePeerConnection(ctx, req)
	//
	//log.Println("req", req.String())
	//log.Println("resp", resp.String())
	//
	//return resp, nil
}

// 1. Get a sdp
// 2. Open data channel connection
//    - Listen for PeerConnection
func (s *SignalingServer) InitiateSignalingConnection(
	ctx context.Context,
	req *pb.InitiatePeerConnectionRequest,
) (res *pb.InitiatePeerConnectionResponse, err error) {

	// Retrieve offer
	offer := webrtc.SessionDescription{}
	b, err := base64.StdEncoding.DecodeString(req.Sdp)
	if err != nil {
		return nil, twirp.NewError(twirp.InvalidArgument, "invalid base64")
	}
	err = json.Unmarshal(b, &offer)
	if err != nil {
		return nil, twirp.NewError(twirp.InvalidArgument, "invalid json")
	}

	log.Println("req", req)
	log.Println("offer", offer)

	answer, err := s.connectToProxy(offer)
	if err != nil {
		log.Println(err)
		return nil, twirp.NewError(twirp.Internal, "cannot generate answer")
	}

	b, err = json.Marshal(answer)
	if err != nil {
		return nil, twirp.NewError(twirp.Internal, "could not generate SDP")
	}

	resp := &pb.InitiatePeerConnectionResponse{
		Sdp: base64.StdEncoding.EncodeToString(b),
	}

	log.Println("req", req.String())
	log.Println("resp", resp.String())

	return resp, nil

}

func (s *SignalingServer) connectToProxy(offer webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	// Create a SettingEngine and enable Detach
	settings := webrtc.SettingEngine{}
	settings.DetachDataChannels()
	webrtcApi := webrtc.NewAPI(
		webrtc.WithSettingEngine(settings))

	peerConnection, err := webrtcApi.NewPeerConnection(webrtc.Configuration{
		// No STUN servers for now, to ensure candidate pair that's selected communicates over LAN
		ICEServers: []webrtc.ICEServer{},
		//ICEServers: []webrtc.ICEServer{
		//	{
		//		URLs: []string{"stun:stun.l.google.com:19302"},
		//	},
		//},
	})

	if err != nil {
		log.Fatal(err)
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Printf("ICE Connection State has changed: %s\n", connectionState.String())
	})

	// Register data channel creation handling
	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		log.Printf("New DataChannel %s %d\n", d.Label(), d.ID())

		// Register channel opening handling
		d.OnOpen(func() {
			log.Printf("Data channel openned; label=%s id=%d  \n", d.Label(), d.ID())

			// Detach the data channel
			raw, dErr := d.Detach()
			if dErr != nil {
				panic(dErr)
			}

			s.raw = &raw

			//log.Println(&raw)
			//log.Println(s.raw)
			//
			//
			//log.Println(raw)
			//log.Println(*s.raw)

			//// Handle reading from the data channel
			//go ReadLoop(raw)
			//
			//// Handle writing to the data channel
			//go WriteLoop(raw)

		})

	})

	// Set the remote SessionDescription
	if err = peerConnection.SetRemoteDescription(offer); err != nil {
		log.Println(err)
		return nil, err
	}
	// Create answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Set the LocalDescription
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		return nil, err
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	return peerConnection.LocalDescription(), nil
}

// ReadLoop shows how to read from the datachannel directly
func (s *SignalingServer) ReadLoop(d io.Reader) {
	for {
		buffer := make([]byte, signalingMessageSizeMax)
		n, err := d.Read(buffer)
		if err != nil {
			log.Println("Datachannel closed; Exit the readloop:", err)
			return
		}

		log.Printf("Message from DataChannel: %s\n", string(buffer[:n]))
	}
}

//// WriteLoop shows how to write to the datachannel directly
//func (s *SignalingServer) WriteLoop(d io.Writer) {
//	for range time.NewTicker(5 * time.Second).C {
//		message := strconv.Itoa(rand.Int())
//		log.Printf("Sending %s \n", message)
//
//		_, err := d.Write([]byte(message))
//		if err != nil {
//			panic(err)
//		}
//	}
//}

