package eventghostadapter

import (
	"fmt"
	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/pkg/eventghostnetwork"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
	"log"
	"strings"
)

type EgReq struct {
	Event   string
	Payload []string
}

type DeviceConn struct {
	Requests chan EgReq
}

func Start(adapter *hapitypes.Adapter, stop *stopper.Stopper) error {
	workers := stopper.NewManager()

	passwordToDeviceId, clientConns, err := startup(adapter, workers)
	if err != nil {
		return err
	}

	go runServer(adapter, passwordToDeviceId, workers)

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

	go func() {
		defer stop.Done()
		defer workers.StopAllWorkersAndWait()

		adapter.Logl.Info.Println("started")
		defer adapter.Logl.Info.Println("stopped")

		for {
			select {
			case <-stop.Signal:
				return
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
	}()

	return nil
}

func runServer(adapter *hapitypes.Adapter, passwordToDeviceId map[string]string, workers *stopper.Manager) {
	passwords := []string{}
	for password, _ := range passwordToDeviceId {
		passwords = append(passwords, password)
	}

	if len(passwords) == 0 {
		adapter.Logl.Info.Println("no EventGhost devices configured - not starting server")
		return
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

	err := eventghostnetwork.RunServer(
		passwords,
		eventHandler,
		workers.Stopper())
	if err != nil {
		adapter.Logl.Error.Println(err.Error())
	}
}

func startup(adapter *hapitypes.Adapter, workers *stopper.Manager) (map[string]string, map[string]*DeviceConn, error) {
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

			go handleClientConnection(
				device.EventghostAddr,
				device.EventghostSecret,
				clientConns[device.DeviceId],
				logex.Prefix(device.EventghostAddr, adapter.Log),
				workers.Stopper())
		}

		passwordToDeviceId[device.EventghostSecret] = device.DeviceId
	}

	return passwordToDeviceId, clientConns, nil
}

func handleClientConnection(
	addr string,
	password string,
	reqs *DeviceConn,
	logger *log.Logger,
	stop *stopper.Stopper,
) {
	defer stop.Done()

	logl := logex.Levels(logger)

	conn := eventghostnetwork.NewEventghostConnection(
		addr,
		password,
		logger)

	for {
		select {
		case req := <-reqs.Requests:
			if err := conn.Send(req.Event, req.Payload); err != nil {
				logl.Error.Println(err.Error())
			}
		case <-stop.Signal:
			return
		}
	}
}
