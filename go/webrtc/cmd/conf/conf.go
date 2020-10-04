package conf

import (
	"fmt"
	"log"
)

const (

	EventBusAddr = "239.20.20.21"
	EventBusPort = 10000
	//RtpAddr      = "127.0.0.1:5000"
	//// Set this too low and we see packet loss in chrome://webrtc-internals, and on the network interface (`netstat -suna`)
	//// But what should it be? `sysctl net.core.rmem_max`?
	//rtpReadBufferSize  = 1024 * 1024 * 8
	//maxRtpDatagramSize = 4096
	//defaultServerAddr  = ":8585"
)

func init() {
	fmt.Println("Init: cmd/conf/conf.go")
	log.SetFlags(log.LstdFlags | log.Llongfile)
}


