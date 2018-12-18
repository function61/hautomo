package zigbee2mqttadapter

import (
	"errors"
	"fmt"
	"github.com/function61/gokit/logger"
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
	"strings"
	"sync"
	"time"
)

var log = logger.New("zigbee2mqtt")

func Start(adapter *hapitypes.Adapter, stop *stopper.Stopper) error {
	// this logic is still TODO and has to be made configurable
	clickRecognizer := func(topicName, message []byte) {
		if string(topicName) != "zigbee2mqtt/0x00158d000227a73c" || !strings.Contains(string(message), `"click":"single"`) {
			return
		}

		log.Info("clicked")

		toggleCabinetLights := hapitypes.NewPowerToggleEvent("45e6e09c")

		for i := 0; i < 6; i++ {
			adapter.Inbound.Receive(&toggleCabinetLights)

			time.Sleep(500 * time.Millisecond)
		}
	}

	subStoppers := stopper.NewManager()

	go func() {
		defer stop.Done()
		defer log.Info("stopped")

		log.Info("started")

		for {
			select {
			case <-stop.Signal:
				subStoppers.StopAllWorkersAndWait()
				return
			case genericEvent := <-adapter.Outbound:
				adapter.LogUnsupportedEvent(genericEvent, log)
			}
		}

	}()

	go func(stop *stopper.Stopper) {
		defer stop.Done()
		defer log.Info("reconnect loop stopped")

		for {
			if err := mqttConnection(adapter.Conf.Zigbee2MqttAddr, clickRecognizer, stop); err != nil {
				log.Error(fmt.Sprintf("mqttConnection error; reconnecting soon: %v", err))
				time.Sleep(1 * time.Second)
			}

			// no error => break
			if stop.SignalReceived {
				return
			}
		}
	}(subStoppers.Stopper())

	return nil
}

func mqttConnection(addr string, handler client.MessageHandler, stop *stopper.Stopper) error {
	broken := make(chan interface{})

	var closeOnce sync.Once

	mqttClient := client.New(&client.Options{
		ErrorHandler: func(err error) {
			fmt.Println(err)

			// TODO: is connection broken now?
			closeOnce.Do(func() {
				close(broken)
			})
		},
	})
	defer mqttClient.Terminate()

	if err := mqttClient.Connect(&client.ConnectOptions{
		Network:  "tcp",
		Address:  addr,
		ClientID: []byte("home-automation-hub"),
	}); err != nil {
		return err
	}

	if err := mqttClient.Subscribe(&client.SubscribeOptions{
		SubReqs: []*client.SubReq{
			{
				TopicFilter: []byte("zigbee2mqtt/#"),
				QoS:         mqtt.QoS0,
				Handler:     handler,
			},
		},
	}); err != nil {
		return err
	}

	select {
	case <-broken:
		return errors.New("connection broken")
	case <-stop.Signal:
		mqttClient.Disconnect()
		return nil
	}
}
