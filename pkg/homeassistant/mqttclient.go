package homeassistant

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/function61/gokit/log/logex"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
)

const (
	ClientIdA = "Hautomo-Home-Assistant1"
	ClientIdB = "Hautomo-EZHub"
)

type InboundCommand struct {
	EntityId string
	Payload  string
}

// sends data to Home Assistant over a MQTT broker
type MqttClient struct {
	mqtt *client.Client
	logl *logex.Leveled
}

func NewMqttClient(
	mqttAddr string,
	clientId string,
	logl *logex.Leveled,
) (*MqttClient, error) {
	mqttClient := client.New(&client.Options{
		ErrorHandler: func(err error) {
			logl.Error.Printf("mqtt: %s", err)
		},
	})

	if err := mqttClient.Connect(&client.ConnectOptions{
		Network:  "tcp",
		Address:  mqttAddr,
		ClientID: []byte(clientId),
	}); err != nil {
		return nil, err
	}

	ha := &MqttClient{
		mqtt: mqttClient,
		logl: logl,
	}

	return ha, nil
}

func (h *MqttClient) SubscribeForCommands(prefix TopicPrefix) (<-chan InboundCommand, error) {
	inboundCh := make(chan InboundCommand, 1)

	if err := h.mqtt.Subscribe(&client.SubscribeOptions{
		SubReqs: []*client.SubReq{
			{
				TopicFilter: []byte(prefix.WildcardCommandSubscriptionPattern()),
				QoS:         mqtt.QoS0,
				Handler: func(topicName, message []byte) {
					inboundCh <- InboundCommand{
						EntityId: prefix.ExtractEntityID(string(topicName)),
						Payload:  string(message),
					}
				},
			},
		},
	}); err != nil {
		return nil, err
	}

	return inboundCh, nil
}

func (h *MqttClient) Close() error {
	h.mqtt.Terminate()
	return nil
}

func (h *MqttClient) PublishState(sensor *Entity, state string) error {
	if err := h.mqtt.Publish(&client.PublishOptions{
		QoS:       mqtt.QoS0,
		Retain:    false,
		TopicName: []byte(sensor.discoveryOpts.StateTopic),
		Message:   []byte(state),
	}); err != nil {
		return fmt.Errorf("PublishState: %w", err)
	}

	return nil
}

func (h *MqttClient) PublishAttributes(entity *Entity, attributes map[string]string) error {
	if len(entity.discoveryOpts.JsonAttributesTopic) == 0 {
		return fmt.Errorf("PublishAttributes: no attribute topic for %s", entity.Id)
	}

	attributesJson, err := json.Marshal(attributes)
	if err != nil {
		return err
	}

	if err := h.mqtt.Publish(&client.PublishOptions{
		QoS:       mqtt.QoS0,
		Retain:    false,
		TopicName: []byte(entity.discoveryOpts.JsonAttributesTopic),
		Message:   attributesJson,
	}); err != nil {
		return fmt.Errorf("PublishAttributes: %w", err)
	}

	return nil
}

// https://www.home-assistant.io/docs/mqtt/discovery/
func (h *MqttClient) AutodiscoverEntities(entities ...*Entity) error {
	for _, entity := range entities {
		if err := h.mqtt.Publish(&client.PublishOptions{
			QoS:       mqtt.QoS0,
			Retain:    true, // so entities are not lost when Home Assistant is restarted
			TopicName: entity.mqttDiscoveryTopic(),
			Message:   entity.mqttDiscoveryMsg(),
		}); err != nil {
			return fmt.Errorf("AutodiscoverSensors: %w", err)
		}
	}

	return nil
}

type TopicPrefix struct {
	prefix string // "hautomo"
}

func NewTopicPrefix(prefix string) TopicPrefix {
	return TopicPrefix{prefix}
}

func (t TopicPrefix) StateTopic(entityId string) string {
	// "hautomo/foobar"
	return fmt.Sprintf("%s/%s", t.prefix, entityId)
}

func (t TopicPrefix) CommandTopic(entityId string) string {
	// "hautomo/foobar/set"
	return fmt.Sprintf("%s/%s/set", t.prefix, entityId)
}

func (t TopicPrefix) AttributesTopic(entityId string) string {
	// "hautomo/foobar/attributes"
	return fmt.Sprintf("%s/%s/attributes", t.prefix, entityId)
}

func (t TopicPrefix) WildcardCommandSubscriptionPattern() string {
	// "hautomo/+/set"
	return fmt.Sprintf("%s/+/set", t.prefix)
}

func (t TopicPrefix) ExtractEntityID(topicName string) string {
	// "hautomo/foobar/set" => "foobar"
	components := strings.Split(topicName, "/")
	if len(components) != 3 {
		return ""
	}

	return components[1]
}
