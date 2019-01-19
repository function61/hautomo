package zigbee2mqttadapter

import (
	"fmt"
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
	"sync"
	"time"
)

const (
	z2mTopicPrefix = "zigbee2mqtt/"
)

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
	config := adapter.GetConfigFileDeprecated()

	resolver := func(adaptersDeviceId string) *resolvedDevice {
		for _, devConfig := range config.Devices {
			// search for incoming messages' device config (same adapter & adapter's device id)
			if devConfig.AdapterId != adapter.Conf.Id || devConfig.AdaptersDeviceId != adaptersDeviceId {
				continue
			}

			kind, found := deviceTypeToZ2mType[devConfig.Type]
			if !found {
				kind = deviceKindUnknown
			}

			return &resolvedDevice{
				id:   devConfig.DeviceId,
				kind: kind,
			}
		}

		return nil
	}

	m2qttDeviceObserver := func(topicName, message []byte) {
		event, err := parseMsgPayload(string(topicName), resolver, string(message))
		if err != nil {
			adapter.Logl.Error.Println(err.Error())
			return
		}

		adapter.Receive(event)
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
		case *hapitypes.BlinkEvent:
			z2mPublish <- deviceMsg(e.DeviceId, `{"alert": "select"}`)
		case *hapitypes.ColorTemperatureEvent:
			deviceConf := config.FindDeviceConfigByAdaptersDeviceId(e.Device)

			deviceType, err := hapitypes.ResolveDeviceType(deviceConf.Type)
			if err != nil {
				panic(err)
			}
			caps := deviceType.Capabilities

			// for some reason IKEA lights with color temp & RGB abilities, do not support
			// color_temp message, so we transparently convert it into a RGB message
			if deviceConf != nil && caps.Color && caps.ColorTemperature {
				r, g, b := temperatureToRGB(float64(e.TemperatureInKelvin))

				// re-publish as a RGB message
				adapter.Outbound <- hapitypes.NewColorMsg(
					e.Device,
					hapitypes.NewRGB(r, g, b))
			} else {
				z2mPublish <- deviceMsg(e.Device, fmt.Sprintf(
					`{"color_temp": %d, "transition": 1}`,
					kelvinToMired(e.TemperatureInKelvin)))
			}
		default:
			adapter.LogUnsupportedEvent(genericEvent)
		}
	}

	go func() {
		defer stop.Done()

		adapter.Logl.Info.Println("started")
		defer adapter.Logl.Info.Println("stopped")

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
		defer adapter.Logl.Info.Println("reconnect loop stopped")

		for {
			if err := mqttConnection(adapter.Conf.Zigbee2MqttAddr, m2qttDeviceObserver, z2mPublish, stop); err != nil {
				adapter.Logl.Error.Printf("mqttConnection error; reconnecting soon: %v", err)
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
				TopicFilter: []byte(z2mTopicPrefix + "#"), // # means catch-all
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
