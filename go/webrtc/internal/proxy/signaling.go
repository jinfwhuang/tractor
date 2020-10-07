package proxy

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	pb "github.com/farm-ng/tractor/genproto"
	"github.com/pion/datachannel"
	"github.com/pion/webrtc/v3"
	"github.com/twitchtv/twirp"
	"google.golang.org/protobuf/proto"
	"log"
	"net/http"
)

const (
	SignalingChanName = "signaling_data_channel"
	SignalingMessageSizeMax = 100000
)



type SignalingConn struct {
	Endpoint string
	pc *webrtc.PeerConnection
	raw *datachannel.ReadWriteCloser
}

func (s *SignalingConn) handleSignalingEvent(event *pb.WebRtcConnection) {
	eventBytes, err := proto.Marshal(event)
	//
	//d := *s.raw
	//_, err = d.Write(eventBytes)

	log.Println("event size", len(eventBytes))

	if err != nil {
		log.Fatal(err)
	}
}

// ReadLoop shows how to read from the datachannel directly
func (s *SignalingConn) ReadLoop() {
	for {
		buffer := make([]byte, SignalingMessageSizeMax)
		_, err := (*s.raw).Read(buffer)
		if err != nil {
			log.Println("Datachannel closed; Exit the readloop:", err)
			return
		}

		event := &pb.WebRtcConnection{}

		proto.Unmarshal(buffer, event)

		log.Println(event)
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




// ------------------------------------------------
// Proxy to Signal
// ------------------------------------------------


// Open a data channel to signaling server
func (s *SignalingConn) ConnectToSignal() error {
	log.Println("starting to connect to signal")
	log.Println("endpoint", s.Endpoint)

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

	// Create a datachannel
	dataChannel, err := peerConnection.CreateDataChannel(SignalingChanName, nil)
	if err != nil {
		panic(err)
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Printf("ICE Connection State has changed: %s\n", connectionState.String())
	})

	// Register channel opening handling
	dataChannel.OnOpen(func() {
		fmt.Printf("Data channel '%s'-'%d' open.\n", dataChannel.Label(), dataChannel.ID())

		// Detach the data channel
		raw, dErr := dataChannel.Detach()
		if dErr != nil {
			panic(dErr)
		}
		s.raw = &raw

		// Handle reading from the data channel
		go s.ReadLoop()
		//go ReadLoop(raw)

		//// Handle writing to the data channel
		//go WriteLoop(raw)
	})

	//// Register channel opening handling
	//dataChannel.OnOpen(func() {
	//	log.Printf(
	//		"Data channel '%s'-'%d' open. Message is sent every 5 seconds\n",
	//		dataChannel.Label(),
	//		dataChannel.ID())
	//
	//	for range time.NewTicker(5 * time.Second).C {
	//		message := "from proxy to signal"
	//		log.Printf("Sending '%s'\n", message)
	//
	//		// Send the message as text
	//		sendErr := dataChannel.SendText(message)
	//		if sendErr != nil {
	//			panic(sendErr)
	//		}
	//	}
	//})

	//// Register text message handling
	//dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
	//	log.Printf("Message received on Proxy server '%s': '%s'\n", dataChannel.Label(), string(msg.Data))
	//})

	// Create an offer to send to the browser
	localSess, err := peerConnection.CreateOffer(nil)
	if err != nil {
		panic(err)
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(localSess)
	if err != nil {
		panic(err)
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	// Communicate with the Singal server
	remoteSess, err := findSignalingPeer(&localSess, s.Endpoint)

	// Apply the answer as the remote description
	err = peerConnection.SetRemoteDescription(*remoteSess)
	if err != nil {
		panic(err)
	}

	log.Println("finished setting up datachannel")

	return nil
}

func findSignalingPeer(
	localSess *webrtc.SessionDescription,
	endpoint string,
	) (*webrtc.SessionDescription, error) {
	sdp, err := json.Marshal(localSess)
	if err != nil {
		log.Fatal(twirp.NewError(twirp.Internal, "could not generate SDP"))
	}
	// Twirp communication
	req := &pb.InitiatePeerConnectionRequest{
		Sdp: base64.StdEncoding.EncodeToString(sdp),
	}

	log.Println("endpoint", endpoint)
	log.Println("making twirp requests", req)

	client := pb.NewWebRTCSignalingServiceJSONClient(endpoint, &http.Client{})
	resp, err := client.InitiateSignalingConnection(context.Background(), req)

	if resp == nil {
		log.Fatal("null response from twirp req")
	}
	log.Println("twirp resp", resp)

	// Deserialize
	remoteSess := webrtc.SessionDescription{}
	sdp, err = base64.StdEncoding.DecodeString(resp.Sdp)
	if err != nil {
		log.Fatal(twirp.NewError(twirp.InvalidArgument, "invalid base64"))
	}
	err = json.Unmarshal(sdp, &remoteSess)
	if err != nil {
		log.Fatal(twirp.NewError(twirp.InvalidArgument, "invalid json"))
	}

	return &remoteSess, nil
}
