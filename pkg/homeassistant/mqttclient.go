package homeassistant

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/function61/gokit/log/logex"
	. "github.com/function61/hautomo/pkg/builtin"
)

// why the fuck aren't these as constants
const (
	mqttQoS0 = 0 // At most once delivery
	mqttQoS1 = 1 // At least once delivery
	mqttQoS2 = 2 // Exactly once delivery
)

type MQTTConfig struct {
	Address     string                 `json:"address"`
	Credentials *MQTTConfigCredentials `json:"credentials"`
}

func (m MQTTConfig) Valid() error {
	return ErrorIfUnset(m.Address == "", "Address")
}

type MQTTConfigCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type task func(context.Context) error

type InboundCommand struct {
	EntityId string
	Payload  string
}

type outgoingItem struct {
	Topic    string
	QoS      byte
	Retained bool
	Payload  []byte
	result   chan<- error // due to re-tries, this is expected to probably always return nil error
}

type subscription struct {
	Topic   string
	QoS     byte
	Handler mqtt.MessageHandler
}

// sends data to Home Assistant over a MQTT broker
type MqttClient struct {
	subReqs  []*subscription
	outgoing []*outgoingItem
	updates  chan Void
	logl     *logex.Leveled
}

func NewMQTTClient(
	mqttConf MQTTConfig,
	clientId string,
	logl *logex.Leveled,
) (*MqttClient, task) {
	ha := &MqttClient{
		subReqs:  []*subscription{},
		outgoing: []*outgoingItem{},
		updates:  make(chan Void, 1),
		logl:     logl,
	}

	return ha, func(ctx context.Context) error {
		if err := mqttConf.Valid(); err != nil {
			return err
		}

		for {
			if err := mqttConnection(ctx, ha, mqttConf, clientId); err != nil {
				logl.Error.Printf("mqttConnection: %v", err)
			}

			select {
			case <-ctx.Done(): // was asked to stop
				return nil
			default:
				time.Sleep(1 * time.Second)
			}
		}
	}
}

func mqttConnection(
	ctx context.Context,
	h *MqttClient,
	conf MQTTConfig,
	clientId string,
) error {
	connectionLost := make(chan error, 1)

	opts := mqtt.NewClientOptions().AddBroker("tcp://" + conf.Address)
	opts.SetClientID(clientId)
	opts.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
		connectionLost <- err
	})
	// we manage reconnects ourselves. it's easier than trying to understand / trust how
	// *ResumeSubs*, *CleanSession* and delivery guarantees of Publish() when reconnecting is needed.
	opts.SetAutoReconnect(false)
	if conf.Credentials != nil {
		opts.SetUsername(conf.Credentials.Username)
		opts.SetPassword(conf.Credentials.Password)
	}

	mqttClient := mqtt.NewClient(opts)
	if err := waitFor(mqttClient.Connect()); err != nil {
		return err
	}
	defer mqttClient.Disconnect(0)

	// handle pending subscriptions (probably has subscriptions only when reconnecting)
	h.connectionStateUpdated()

	// subscriptions are per-connection. if we disconnect, we want to subscribe to the same
	// topics again that we had during the previous connection's lifecycle
	subscribedIdx := 0

	for {
		select {
		case <-ctx.Done(): // was asked to stop
			return nil
		case err := <-connectionLost:
			return err
		case <-h.updates:
			toSubscribe := h.subReqs[subscribedIdx:]
			for _, subscription := range toSubscribe {
				if err := waitFor(mqttClient.Subscribe(subscription.Topic, subscription.QoS, subscription.Handler)); err != nil {
					return err
				}

				subscribedIdx++
			}

			for len(h.outgoing) > 0 { // send all outgoing
				outgoing := h.outgoing[0]

				if err := waitFor(mqttClient.Publish(outgoing.Topic, outgoing.QoS, outgoing.Retained, outgoing.Payload)); err != nil {
					return err
				}

				outgoing.result <- nil // succeeded

				h.outgoing = h.outgoing[1:]
			}
		}
	}
}

func (h *MqttClient) SubscribeForCommands(prefix TopicPrefix) (<-chan InboundCommand, error) {
	inboundCh := make(chan InboundCommand, 1)

	h.subReqs = append(h.subReqs, &subscription{
		Topic: prefix.WildcardCommandSubscriptionPattern(),
		QoS:   mqttQoS0,
		Handler: func(_ mqtt.Client, message mqtt.Message) {
			inboundCh <- InboundCommand{
				EntityId: prefix.ExtractEntityID(message.Topic()),
				Payload:  string(message.Payload()),
			}
		},
	})

	h.connectionStateUpdated()

	return inboundCh, nil
}

func (h *MqttClient) PublishState(sensor *Entity, state string) <-chan error {
	return h.out(&outgoingItem{
		Topic:    sensor.discoveryOpts.StateTopic,
		QoS:      mqttQoS0,
		Retained: false,
		Payload:  []byte(state),
	})
}

func (h *MqttClient) PublishTopic(sensor *Entity, payload []byte) <-chan error {
	return h.out(&outgoingItem{
		Topic:    sensor.discoveryOpts.Topic,
		QoS:      mqttQoS0,
		Retained: false,
		Payload:  payload,
	})
}

func (h *MqttClient) PublishAttributes(entity *Entity, attributes map[string]interface{}) <-chan error {
	if len(entity.discoveryOpts.JsonAttributesTopic) == 0 {
		return immediateError(
			fmt.Errorf("PublishAttributes: no attribute topic for %s", entity.Id))
	}

	attributesJson, err := json.Marshal(attributes)
	if err != nil {
		return immediateError(err)
	}

	return h.out(&outgoingItem{
		Topic:    entity.discoveryOpts.JsonAttributesTopic,
		QoS:      mqttQoS0,
		Retained: false,
		Payload:  attributesJson,
	})
}

// https://www.home-assistant.io/docs/mqtt/discovery/
func (h *MqttClient) AutodiscoverEntities(entities ...*Entity) error {
	// publishes many messages, so we'll block to find out results because it's harder
	// to return future<error> concerning many pending operations
	for _, entity := range entities {
		if err := <-h.out(&outgoingItem{
			Topic:    string(entity.mqttDiscoveryTopic()),
			QoS:      mqttQoS0,
			Retained: true, // so entities are not lost when Home Assistant is restarted
			Payload:  entity.mqttDiscoveryMsg(),
		}); err != nil {
			return fmt.Errorf("AutodiscoverEntities: %w", err)
		}
	}

	return nil
}

func (h *MqttClient) out(msg *outgoingItem) <-chan error {
	result := make(chan error, 1)

	msg.result = result

	// must keep a list for re-tries
	h.outgoing = append(h.outgoing, msg)

	h.connectionStateUpdated()

	return result
}

func (h *MqttClient) connectionStateUpdated() {
	select {
	case h.updates <- Void{}:
	default:
	}
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

func immediateError(err error) <-chan error {
	future := make(chan error, 1)
	future <- err
	return future
}

func waitFor(token mqtt.Token) error {
	<-token.Done()
	return token.Error()
}
