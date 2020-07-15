// Deduces a person's presence by her device (e.g. cell phone). Turns out this is not a
// good idea because aggressive power saving means the devices will miss ping requests
package presencebypingadapter

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/taskrunner"
	"github.com/function61/hautomo/pkg/hapitypes"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
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

func Start(ctx context.Context, adapter *hapitypes.Adapter) error {
	// this is a privileged operation, you need to set:
	// "$ sudo sysctl 'net.ipv4.ping_group_range=0   27'"
	icmpSocket, err := icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		return err
	}
	defer icmpSocket.Close()

	forStamping := make(chan ProbeRequest, 16)
	pingRequests := make(chan ProbeRequest, 16)
	pingResponses := make(chan ProbeResponse, 16)

	tasks := taskrunner.New(ctx, adapter.Log)

	tasks.Start("tickerLoop", func(ctx context.Context) error {
		return tickerLoop(ctx, adapter.Conf, adapter, forStamping, pingResponses)
	})

	tasks.Start("pingSender", func(ctx context.Context) error {
		return pingSender(ctx, icmpSocket, pingRequests, adapter.Logl)
	})

	tasks.Start("pingReceiver", func(ctx context.Context) error {
		return pingReceiver(ctx, icmpSocket, pingResponses, adapter.Logl)
	})

	inFlight := map[int]ProbeRequest{}

	for {
		select {
		case <-ctx.Done():
			return tasks.Wait()
		case err := <-tasks.Done(): // subtask crash
			return err
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
	ctx context.Context,
	config hapitypes.AdapterConfig,
	adapter *hapitypes.Adapter,
	forStamping chan<- ProbeRequest,
	pingResponses chan<- ProbeResponse,
) error {
	personIdPresentMap := map[string]bool{}

	probeCount := len(config.PresenceByPingDevice)

	presences := make(chan Presence, probeCount)

	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil
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
	ctx context.Context,
	icmpSocket *icmp.PacketConn,
	pingResponses chan<- ProbeResponse,
	logl *logex.Leveled,
) error {
	go func() {
		<-ctx.Done()
		icmpSocket.Close() // only way to unblock below ReadFrom()
	}()

	for {
		readBuffer := make([]byte, 1500)

		bytesRead, peer, err := icmpSocket.ReadFrom(readBuffer)
		if err != nil {
			select {
			case <-ctx.Done(): // expected error
				return nil
			default: // unexpected error, probably cannot read from socket anymore
				return err
			}
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
			// TODO: what's with the plus v?
			logl.Debug.Printf("pingReceiver: got %+v; want echo reply", icmpMsg)
		}
	}
}

func pingSender(
	ctx context.Context,
	icmpSocket *icmp.PacketConn,
	requests <-chan ProbeRequest,
	logl *logex.Leveled,
) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case req := <-requests:
			echoRequest := icmp.Message{
				Type: ipv4.ICMPTypeEcho,
				Code: 0,
				Body: &req.PingPacket,
			}

			echoRequestBytes, err := echoRequest.Marshal(nil)
			if err != nil { // should not happen
				return fmt.Errorf("echoRequest.Marshal: %w", err)
			}

			logl.Debug.Printf("pingSender: %s", req.IP.String())

			if _, err := icmpSocket.WriteTo(echoRequestBytes, &net.UDPAddr{
				IP: req.IP,
			}); err != nil {
				// probably cannot write to socket anymore, so stop whole operation
				return err
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
