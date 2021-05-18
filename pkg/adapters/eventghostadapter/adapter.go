package eventghostadapter

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/function61/gokit/log/logex"
	"github.com/function61/gokit/sync/taskrunner"
	"github.com/function61/hautomo/pkg/eventghostnetwork"
	"github.com/function61/hautomo/pkg/hapitypes"
)

type EgReq struct {
	Event   string
	Payload []string
}

type DeviceConn struct {
	Requests chan EgReq
}

func Start(ctx context.Context, adapter *hapitypes.Adapter) error {
	serverAndClientConnections := taskrunner.New(ctx, adapter.Log)

	// connect client to all computers we wish to send EventGhost commands to
	passwordToDeviceId, clientConns, err := readConfAndConnectToClients(adapter, func(eventghostAddr string, eventghostSecret string, deviceConn *DeviceConn) {
		serverAndClientConnections.Start(eventghostAddr, func(ctx context.Context) error {
			return handleClientConnection(
				ctx,
				eventghostAddr,
				eventghostSecret,
				deviceConn,
				logex.Prefix(eventghostAddr, adapter.Log))
		})
	})
	if err != nil {
		return err
	}

	// start server to receive EventGhost-sent events to Hautomo
	serverAndClientConnections.Start("server", func(ctx context.Context) error {
		return runServer(ctx, adapter, passwordToDeviceId)
	})

	send := func(deviceId string, req EgReq) {
		conn, found := clientConns[deviceId]
		if !found {
			adapter.Logl.Error.Printf("unknown device %s", deviceId)
			return
		}

		select {
		case conn.Requests <- req:
		default:
			adapter.Logl.Error.Printf("device %s queue is full; message discarded", deviceId)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return serverAndClientConnections.Wait()
		case err := <-serverAndClientConnections.Done(): // subtask crash
			return err
		case genericEvent := <-adapter.Outbound:
			switch e := genericEvent.(type) {
			case *hapitypes.NotificationEvent:
				send(e.Device, EgReq{
					Event:   "OSD",
					Payload: []string{e.Message},
				})
			case *hapitypes.PlaybackEvent:
				send(e.Device, EgReq{
					Event: "Playback." + e.Action,
				})
			default:
				adapter.LogUnsupportedEvent(genericEvent)
			}
		}
	}
}

func runServer(
	ctx context.Context,
	adapter *hapitypes.Adapter,
	passwordToDeviceId map[string]string,
) error {
	passwords := []string{}
	for password := range passwordToDeviceId {
		passwords = append(passwords, password)
	}

	if len(passwords) == 0 {
		adapter.Logl.Info.Println("no EventGhost devices configured - not starting server")
		return nil
	}

	eventHandler := func(event string, payload []string, password string) {
		payloadSerialized := ""
		if len(payload) > 0 {
			payloadSerialized = ":" + strings.Join(payload, ":")
		}

		deviceId := passwordToDeviceId[password]

		// TODO: model "un-idle" event coming from PC as a structural "idle detector" event OR movement sensor event?
		adapter.Receive(hapitypes.NewPublishEvent("eventghost:" + deviceId + ":" + event + payloadSerialized))
	}

	return eventghostnetwork.RunServer(
		ctx,
		passwords,
		eventHandler)
}

func readConfAndConnectToClients(
	adapter *hapitypes.Adapter,
	connect func(string, string, *DeviceConn),
) (map[string]string, map[string]*DeviceConn, error) {
	passwordToDeviceId := map[string]string{}
	clientConns := map[string]*DeviceConn{}

	conf := adapter.GetConfigFileDeprecated()
	for _, device := range conf.Devices {
		if device.AdapterId != adapter.Conf.Id {
			continue
		}

		if device.EventghostSecret == "" {
			return nil, nil, fmt.Errorf("empty EventghostSecret")
		}

		if _, duplicatePassword := passwordToDeviceId[device.EventghostSecret]; duplicatePassword {
			return nil, nil, fmt.Errorf("duplicate EventghostSecret detected")
		}

		if device.EventghostAddr != "" {
			clientConns[device.DeviceId] = &DeviceConn{
				Requests: make(chan EgReq, 16),
			}

			connect(device.EventghostAddr, device.EventghostSecret, clientConns[device.DeviceId])
		}

		passwordToDeviceId[device.EventghostSecret] = device.DeviceId
	}

	return passwordToDeviceId, clientConns, nil
}

func handleClientConnection(
	ctx context.Context,
	addr string,
	password string,
	reqs *DeviceConn,
	logger *log.Logger,
) error {
	logl := logex.Levels(logger)

	// internally manages reconnects
	conn := eventghostnetwork.NewEventghostConnection(
		addr,
		password,
		logger)

	for {
		select {
		case <-ctx.Done():
			return nil
		case req := <-reqs.Requests:
			if err := conn.Send(req.Event, req.Payload); err != nil {
				logl.Error.Println(err.Error())
			}
		}
	}
}
