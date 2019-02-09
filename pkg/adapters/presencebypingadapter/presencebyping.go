package presencebypingadapter

import (
	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/stopper"
	"github.com/function61/hautomo/pkg/hapitypes"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"net"
	"sync"
	"time"
)

type ProbeRequest struct {
	PingPacket icmp.Echo
	IP         net.IP
	GotReply   chan interface{} // closed when got back reply
}

type ProbeResponse struct {
	ID      int
	Timeout bool
}

type Presence struct {
	Person  string
	Present bool
}

func Start(adapter *hapitypes.Adapter, stop *stopper.Stopper) error {
	workers := stopper.NewManager()

	// this is a privileged operation, you need to set:
	// "$ sudo sysctl 'net.ipv4.ping_group_range=0   27'"
	icmpSocket, err := icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		stop.Done() // TODO: have this done robustly?
		return err
	}

	forStamping := make(chan ProbeRequest, 16)
	pingRequests := make(chan ProbeRequest, 16)
	pingResponses := make(chan ProbeResponse, 16)

	go tickerLoop(adapter.Conf, adapter, forStamping, pingResponses, workers.Stopper())

	go pingSender(icmpSocket, pingRequests, adapter.Logl, workers.Stopper())

	go pingReceiver(icmpSocket, pingResponses, adapter.Logl, workers.Stopper())

	go func(stop *stopper.Stopper) {
		defer stop.Done()

		inFlight := map[int]ProbeRequest{}

		for {
			select {
			case <-stop.Signal:
				return
			case in := <-pingResponses:
				if input, has := inFlight[in.ID]; has {
					close(input.GotReply)
					delete(inFlight, in.ID)
				}
			case out := <-forStamping:
				inFlight[out.PingPacket.ID] = out

				pingRequests <- out
			}
		}
	}(workers.Stopper())

	go func() {
		defer icmpSocket.Close()
		defer stop.Done()
		<-stop.Signal
		workers.StopAllWorkersAndWait()
	}()

	return nil
}

func probePresence(
	pbpd hapitypes.PresenceByPingDevice,
	forStamping chan<- ProbeRequest,
	pingResponses chan<- ProbeResponse,
	presences chan<- Presence,
) {
	probe := ProbeRequest{
		PingPacket: icmp.Echo{
			ID:   uniquePingId(),
			Seq:  1,
			Data: []byte("HELLO-R-U-THERE"),
		},
		IP:       net.ParseIP(pbpd.Ip),
		GotReply: make(chan interface{}),
	}

	reportPresent := func(present bool) {
		presences <- Presence{
			Person:  pbpd.Person,
			Present: present,
		}
	}

	forStamping <- probe

	select {
	case <-probe.GotReply:
		reportPresent(true)
	case <-time.After(1 * time.Second): // timeout
		pingResponses <- ProbeResponse{
			ID:      probe.PingPacket.ID,
			Timeout: true,
		}

		reportPresent(false)
	}
}

func tickerLoop(
	config hapitypes.AdapterConfig,
	adapter *hapitypes.Adapter,
	forStamping chan<- ProbeRequest,
	pingResponses chan<- ProbeResponse,
	stop *stopper.Stopper,
) {
	defer stop.Done()

	personIdPresentMap := map[string]bool{}

	probeCount := len(config.PresenceByPingDevice)

	presences := make(chan Presence, probeCount)

	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-stop.Signal:
			return
		case <-ticker.C:
			// launch these in parallel
			for _, pbpd := range config.PresenceByPingDevice {
				go probePresence(pbpd, forStamping, pingResponses, presences)
			}

			for i := 0; i < probeCount; i++ {
				current := <-presences

				previous, firstResult := personIdPresentMap[current.Person]

				// TODO: have this "different than last one" -check in app
				if !firstResult || current.Present != previous {
					adapter.Receive(hapitypes.NewPersonPresenceChangeEvent(
						current.Person,
						current.Present))
				}

				personIdPresentMap[current.Person] = current.Present
			}
		}
	}
}

func pingReceiver(
	icmpSocket *icmp.PacketConn,
	pingResponses chan<- ProbeResponse,
	logl *logex.Leveled,
	stop *stopper.Stopper,
) {
	defer stop.Done()
	defer logl.Info.Println("pingReceiver stopped")

	go func() {
		<-stop.Signal
		icmpSocket.Close() // only way to unblock below ReadFrom()
	}()

	for {
		readBuffer := make([]byte, 1500)

		bytesRead, peer, err := icmpSocket.ReadFrom(readBuffer)
		if err != nil {
			logl.Error.Printf("pingReceiver: stopping due to %s", err.Error())
			// probably cannot read from socket anymore
			return
		}

		// golang.org/x/net/internal/iana
		//     for ipv6 this would be iana.ProtocolIPv6ICMP
		// icmpMsg, err := icmp.ParseMessage(iana.ProtocolICMP, readBuffer[:bytesRead])
		icmpMsg, err := icmp.ParseMessage(1, readBuffer[:bytesRead])
		if err != nil {
			logl.Error.Printf("pingReceiver: ParseMessage: %s", err.Error())
			continue
		}

		switch icmpMsg.Type {
		case ipv4.ICMPTypeEchoReply:
			echoReply := icmpMsg.Body.(*icmp.Echo)

			logl.Debug.Printf("pingReceiver: got reply from %v", peer)

			pingResponses <- ProbeResponse{
				ID:      echoReply.ID,
				Timeout: false,
			}
		default:
			logl.Debug.Printf("pingReceiver: got %+v; want echo reply", icmpMsg)
		}
	}
}

func pingSender(
	icmpSocket *icmp.PacketConn,
	requests <-chan ProbeRequest,
	logl *logex.Leveled,
	stop *stopper.Stopper,
) {
	defer stop.Done()
	defer logl.Info.Println("pingSender stopped")

	for {
		select {
		case <-stop.Signal:
			return
		case req := <-requests:
			echoRequest := icmp.Message{
				Type: ipv4.ICMPTypeEcho,
				Code: 0,
				Body: &req.PingPacket,
			}

			echoRequestBytes, err := echoRequest.Marshal(nil)
			if err != nil {
				logl.Error.Printf("pingSender: %s", err.Error())
				return
			}

			logl.Debug.Printf("pingSender: %s", req.IP.String())

			if _, err := icmpSocket.WriteTo(echoRequestBytes, &net.UDPAddr{
				IP: req.IP,
			}); err != nil {
				logl.Error.Printf("pingSender: stopping due to %s", err.Error())
				// probably cannot write to socket anymore
				return
			}
		}
	}
}

var pingIdx = 0
var pingIdxMu = sync.Mutex{}

// FIXME: provide this counter from the calling goroutine, therefore avoiding need for mutex
func uniquePingId() int {
	pingIdxMu.Lock()
	defer pingIdxMu.Unlock()

	// WATCH OUT: ID is 16-bit even though it's declared as an int
	if pingIdx >= 0xffff {
		pingIdx = 0
	}

	pingIdx++

	return pingIdx
	// return os.Getpid() & 0xffff
}
