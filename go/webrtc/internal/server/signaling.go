package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	pb "github.com/farm-ng/tractor/genproto"
	"github.com/farm-ng/tractor/webrtc/internal/common"
	"github.com/golang/protobuf/ptypes"
	"github.com/pion/datachannel"
	"github.com/pion/webrtc/v3"
	"github.com/twitchtv/twirp"
	"google.golang.org/protobuf/proto"
	"time"

	//"github.com/pion/webrtc/v3/signal"
	//"github.com/pion/webrtc/v3/examples/internal/signal"
	"log"
)

//type Req struct {
//	Sdp string `json:"sdp"`
//}

const (
	SignalingMessageSizeMax = 1_000_000 // 1mb
)

/*
Assumptions
1. Only talk to a single proxy server
*/
type SignalingServer struct {
	conns map[string]*pb.WebrtcPeerConn

	pc *webrtc.PeerConnection
	raw *datachannel.ReadWriteCloser
}

// ReadLoop shows how to read from the datachannel directly
func (s *SignalingServer) ReadLoop() {
	s.conns = make(map[string]*pb.WebrtcPeerConn)
	for {
		buffer := make([]byte, SignalingMessageSizeMax)
		_, err := (*s.raw).Read(buffer)
		if err != nil {
			log.Println("Datachannel closed; Exit the readloop:", err)
			panic(err)
		}
		fromProxy := &pb.WebrtcPeerConn{}
		proto.Unmarshal(buffer, fromProxy)
		s.conns[fromProxy.ConnId] = fromProxy
	}
}


func (s *SignalingServer) emitSignalingEvent(connId string, req *pb.InitiatePeerConnectionRequest) {
	if s.raw == nil {
		panic("signaling Server has not been properly initialized")
	}

	event := &pb.WebRtcConnection{
		Stamp: ptypes.TimestampNow(),
		ConnId: connId,
		ClientSdp: req.Sdp,
	}
	eventBytes, err := proto.Marshal(event)

	_, err = (*s.raw).Write(eventBytes)
	if err != nil {
		panic(err)
	}
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
			go s.ReadLoop()

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

// InitiatePeerConnection starts the proxy and returns an SDP answer to the client
func (s *SignalingServer) initPeerConnection(
	ctx context.Context,
	req *pb.InitiatePeerConnectionRequest,
) (res *pb.InitiatePeerConnectionResponse, err error) {

	peerId := common.UniquePeerID()
	s.emitSignalingEvent(peerId, req)

	sleepDuration := 300 * time.Millisecond
	reqWaitDuration := 10 * time.Second
	elapsed := 0 * time.Second
	for elapsed < reqWaitDuration {
		time.Sleep(300 * time.Millisecond)
		elapsed += sleepDuration
		if val, exists := s.conns[peerId]; exists {
			return &pb.InitiatePeerConnectionResponse{
				Sdp: val.ProxySdp,
			}, nil
		}
	}
	log.Println("cannot find a peer connection", s.conns)
	return nil, twirp.NewError(twirp.Internal, "cannot find a peer connection")
}

// TODO: It is hacky. It abuses the WebRTCProxyService definition. The convenience of reusing
//       the service is that the frontend app does not need to be configured whether it is
//       talking to API server or SIGNALING server
//
func (s *SignalingServer) InitiatePeerConnection(
	ctx context.Context,
	req *pb.InitiatePeerConnectionRequest,
) (res *pb.InitiatePeerConnectionResponse, err error) {
	return s.initPeerConnection(ctx, req)
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

func (s *SignalingServer) Conns(
	ctx context.Context,
	req *pb.ConnsReq,
) (res *pb.ConnsResponse, err error) {
	resp := &pb.ConnsResponse{
		Conns: s.conns,
	}
	return resp, nil
}

