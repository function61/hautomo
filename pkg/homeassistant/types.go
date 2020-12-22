package homeassistant

import (
	"encoding/json"
	"fmt"
)

const (
	DeviceClassDefault = ""
	DeviceClassMotion  = "motion"
)

type State struct {
	State      string            `json:"state"`
	Attributes map[string]string `json:"attributes"`
}

type Entity struct {
	component          string
	deviceClass        string // https://www.home-assistant.io/integrations/binary_sensor/ | https://www.home-assistant.io/integrations/sensor/
	Id                 string
	name               string
	hasAttributesTopic bool
	hasCommandTopic    bool
}

func NewBinarySensor(id string, name string, deviceClass string) *Entity {
	return &Entity{
		Id:          id,
		name:        name,
		component:   "binary_sensor",
		deviceClass: deviceClass,
	}
}

func NewSwitch(id string, name string) *Entity {
	return &Entity{
		Id:              id,
		name:            name,
		component:       "switch",
		hasCommandTopic: true,
	}
}

func NewSensor(id string, name string, deviceClass string, hasAttributesTopic bool) *Entity {
	return &Entity{
		Id:                 id,
		name:               name,
		component:          "sensor",
		deviceClass:        deviceClass,
		hasAttributesTopic: hasAttributesTopic,
	}
}

func (h *Entity) mqttStateTopic() []byte {
	return []byte(fmt.Sprintf("homeassistant/%s/hautomo/%s/state", h.component, h.Id))
}

func (h *Entity) mqttAttributesTopic() []byte {
	if !h.hasAttributesTopic {
		return nil
	}

	return []byte(fmt.Sprintf("homeassistant/%s/hautomo/%s/attributes", h.component, h.Id))
}

func (h *Entity) mqttCommandTopic() []byte {
	if !h.hasCommandTopic {
		return nil
	}

	return []byte(fmt.Sprintf("homeassistant/%s/hautomo/%s/set", h.component, h.Id))
}

func (h *Entity) mqttDiscoveryTopic() []byte {
	// <discovery_prefix>/<component>/[<node_id>/]<object_id>/config
	return []byte(fmt.Sprintf("homeassistant/%s/hautomo/%s/config", h.component, h.Id))
}

func (h *Entity) mqttDiscoveryMsg() []byte {
	// keys: https://www.home-assistant.io/docs/mqtt/discovery/#configuration-variables
	// many of the "omitempty" attributes are strictly required
	msg, err := json.Marshal(struct {
		Name                string `json:"name"`
		DeviceClass         string `json:"device_class,omitempty"`
		StateTopic          string `json:"state_topic,omitempty"`
		CommandTopic        string `json:"command_topic,omitempty"`
		JsonAttributesTopic string `json:"json_attributes_topic,omitempty"`
		ValueTemplate       string `json:"value_template,omitempty"`
		// UniqueId            string `json:"unique_id,omitempty"`
	}{
		Name:                h.Id,
		DeviceClass:         h.deviceClass,
		StateTopic:          string(h.mqttStateTopic()),
		CommandTopic:        string(h.mqttCommandTopic()),
		JsonAttributesTopic: string(h.mqttAttributesTopic()),
		// UniqueId:            h.Id, // same as entity ID?
	})
	if err != nil {
		panic(err)
	}
	return msg
}
