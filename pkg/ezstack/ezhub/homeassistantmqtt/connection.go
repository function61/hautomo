package homeassistantmqtt

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/function61/gokit/encoding/jsonfile"
	"github.com/function61/hautomo/pkg/ezstack/zigbee"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
)

type Message struct {
	Topic   string
	Content string
}

type InboundMessage struct {
	DeviceId zigbee.IEEEAddress
	Message  zigbee2mqttGenericJson
}

func ConnectAndServe(
	ctx context.Context,
	addr string,
	mqttPrefix string,
	outbound <-chan Message,
	inbound chan<- InboundMessage,
) error {
	// to debug:
	// $ docker run -d --name mosquitto -p 1883:1883 eclipse-mosquitto:1.6.12
	// $ docker exec -it mosquitto mosquitto_sub -t zigbee2mqtt/DEVICE_FRIENDLY_NAME

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

	// we can use mqttPrefix to serve two Zigbee radios on a single node, and thus we must mix it
	// with MQTT client ID because MQTT servers usually allow only one connection per ClientID
	if err := mqttClient.Connect(&client.ConnectOptions{
		Network:  "tcp",
		Address:  addr,
		ClientID: []byte("ezhub-" + mqttPrefix),
	}); err != nil {
		return err
	}

	if err := mqttClient.Subscribe(&client.SubscribeOptions{
		SubReqs: []*client.SubReq{
			{
				TopicFilter: []byte(mqttPrefix + "/+/set"), // + means single-level catch-all
				QoS:         mqtt.QoS0,
				Handler: func(topicName, message []byte) {
					// "joonas/0xec1bbdfffe210132/set" => "0xec1bbdfffe210132"
					address := strings.Split(string(topicName), "/")[1]

					msg := zigbee2mqttGenericJson{}

					// we can't get Home Assistant to both send us a payload of {"state": "ON"} and
					// have it parse it too, so to allow HA to parse state from JSON we must tolerate it
					// sending us non-JSON
					switch string(message) {
					case "OPEN", "CLOSE", "STOP":
						hackShadeCommand := string(message)
						msg.HackShadeCommand = &hackShadeCommand
					case "ON", "OFF":
						messageCopy := string(message) // need copy to get ptr
						msg.State = &messageCopy
					default:
						if err := jsonfile.UnmarshalDisallowUnknownFields(bytes.NewReader(message), &msg); err != nil {
							breakConnectionWithError(fmt.Errorf("not JSON: %s", string(message)))
							return
						}
					}

					inbound <- InboundMessage{
						DeviceId: zigbee.IEEEAddress(address),
						Message:  msg,
					}
				},
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
			case publish := <-outbound:
				if err := mqttClient.Publish(&client.PublishOptions{
					QoS:       mqtt.QoS0,
					Retain:    false,
					TopicName: []byte(publish.Topic),
					Message:   []byte(publish.Content),
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
