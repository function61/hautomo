package zigbee2mqttadapter

import (
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

type MqttPublish struct {
	Topic   string
	Message string
}

func deviceMsg(deviceId string, msg string) MqttPublish {
	return MqttPublish{
		Topic:   "zigbee2mqtt/" + deviceId + "/set",
		Message: msg,
	}
}

func Start(adapter *hapitypes.Adapter, stop *stopper.Stopper) error {
	// this logic is still TODO and has to be made configurable
	clickRecognizer := func(topicName, message []byte) {
		if string(topicName) != "zigbee2mqtt/0x00158d000227a73c" || !strings.Contains(string(message), `"click":"single"`) {
			return
		}

		log.Info("clicked")

		toggleCabinetLights := hapitypes.NewPowerToggleEvent("45e6e09c")

		for i := 0; i < 4; i++ {
			adapter.Inbound.Receive(&toggleCabinetLights)

			time.Sleep(1000 * time.Millisecond)
		}
	}

	z2mPublish := make(chan MqttPublish, 16)

	subStoppers := stopper.NewManager()

	handleOneEvent := func(genericEvent hapitypes.OutboundEvent) {
		switch e := genericEvent.(type) {
		case *hapitypes.PowerMsg:
			if e.On {
				z2mPublish <- deviceMsg(e.DeviceId, `{"state": "ON", "transition": 3}`)
			} else {
				z2mPublish <- deviceMsg(e.DeviceId, `{"state": "OFF", "transition": 3}`)
			}
		case *hapitypes.BrightnessMsg:
			// 0-100 => 0-255
			to := int(float64(e.Brightness) * 2.55)

			z2mPublish <- deviceMsg(e.DeviceId, fmt.Sprintf(`{"brightness": %d, "transition": 2}`, to))
		case *hapitypes.ColorMsg:
			z2mPublish <- deviceMsg(e.DeviceId, fmt.Sprintf(
				`{"color": {"r": %d, "g": %d, "b": %d}, "transition": 2}`,
				e.Color.Red,
				e.Color.Green,
				e.Color.Blue))
		case *hapitypes.ColorTemperatureEvent:
			z2mPublish <- deviceMsg(e.Device, fmt.Sprintf(
				`{"color_temp": %d, "transition": 1}`,
				kelvinToMired(e.TemperatureInKelvin)))
		default:
			adapter.LogUnsupportedEvent(genericEvent, log)
		}
	}

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
				handleOneEvent(genericEvent)
			}
		}
	}()

	go func(stop *stopper.Stopper) {
		defer stop.Done()
		defer log.Info("reconnect loop stopped")

		for {
			if err := mqttConnection(adapter.Conf.Zigbee2MqttAddr, clickRecognizer, z2mPublish, stop); err != nil {
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

func mqttConnection(addr string, handler client.MessageHandler, mqttPublishes <-chan MqttPublish, stop *stopper.Stopper) error {
	broken := make(chan interface{})
	var brokenErr error
	var brokenOnce sync.Once

	// there might be multiple error signals coming in - only process the first.
	breakConnectionWithError := func(err error) {
		brokenOnce.Do(func() {
			brokenErr = fmt.Errorf("mqtt connection broke: %v", err)
			close(broken) // signals total teardown of this connection
		})
	}

	mqttClient := client.New(&client.Options{
		ErrorHandler: func(err error) {
			breakConnectionWithError(err)
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
				TopicFilter: []byte("zigbee2mqtt/#"), // # means catch-all
				QoS:         mqtt.QoS0,
				Handler:     handler,
			},
		},
	}); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-broken:
				return
			case publish := <-mqttPublishes:
				if err := mqttClient.Publish(&client.PublishOptions{
					QoS:       mqtt.QoS0,
					Retain:    false,
					TopicName: []byte(publish.Topic),
					Message:   []byte(publish.Message),
				}); err != nil {
					breakConnectionWithError(err)
					return
				}
			}
		}
	}()

	select {
	case <-broken:
		return brokenErr
	case <-stop.Signal:
		mqttClient.Disconnect()
		return nil
	}
}
