package zigbee2mqttadapter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/function61/gokit/sync/taskrunner"
	"github.com/function61/hautomo/pkg/hapitypes"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
)

type MqttPublish struct {
	Topic   string
	Message string
}

func Start(ctx context.Context, adapter *hapitypes.Adapter) error {
	config := adapter.GetConfigFileDeprecated()

	topicPrefix := adapter.Conf.MqttTopicPrefix
	if topicPrefix == "" {
		topicPrefix = "zigbee2mqtt" // default
	}

	deviceMsg := func(deviceId string, msg string) MqttPublish {
		return MqttPublish{
			Topic:   topicPrefix + "/" + deviceId + "/set",
			Message: msg,
		}
	}

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
		events, err := parseMsgPayload(string(topicName), topicPrefix, resolver, string(message), time.Now())
		if err != nil {
			adapter.Logl.Error.Println(err.Error())
			return
		}

		for _, event := range events {
			adapter.Receive(event)
		}
	}

	z2mPublish := make(chan MqttPublish, 16)

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

	subTasks := taskrunner.New(ctx, adapter.Log)
	subTasks.Start("reconnect-loop", func(ctx context.Context) error {
		for {
			err := mqttConnection(ctx, adapter.Conf.Url, topicPrefix, m2qttDeviceObserver, z2mPublish)

			select {
			case <-ctx.Done():
				return nil
			default:
				adapter.Logl.Error.Printf("mqttConnection error; reconnecting soon: %v", err)
				time.Sleep(1 * time.Second)
			}
		}
	})

	for {
		select {
		case <-ctx.Done():
			return subTasks.Wait()
		case err := <-subTasks.Done(): // subtask crash
			return err
		case genericEvent := <-adapter.Outbound:
			handleOneEvent(genericEvent)
		}
	}
}

func mqttConnection(
	ctx context.Context,
	addr string,
	topicPrefix string,
	handler client.MessageHandler,
	mqttPublishes <-chan MqttPublish,
) error {
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
		ClientID: []byte(fmt.Sprintf("Hautomo-%s", topicPrefix)), // need to mix in topic prefix to support multiple instances
	}); err != nil {
		return err
	}

	if err := mqttClient.Subscribe(&client.SubscribeOptions{
		SubReqs: []*client.SubReq{
			{
				TopicFilter: []byte(topicPrefix + "/#"), // # means catch-all
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
	case <-ctx.Done():
		return mqttClient.Disconnect()
	}
}
