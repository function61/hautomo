package homeassistant

import (
	"encoding/json"
	"fmt"
)

type Component string

const (
	ComponentSwitch       Component = "switch"
	ComponentLight        Component = "light"
	ComponentSensor       Component = "sensor"
	ComponentCover        Component = "cover"
	ComponentBinarySensor Component = "binary_sensor"
)

// NOTE: device classes are platform-specific, i.e. "door" device class is only recognized by
//       binary_sensor platform
const (
	DeviceClassDefault     = ""
	DeviceClassMotion      = "motion"      // component=binary_sensor
	DeviceClassOccupancy   = "occupancy"   // component=binary_sensor
	DeviceClassDoor        = "door"        // component=binary_sensor
	DeviceClassTemperature = "temperature" // component=sensor
	DeviceClassHumidity    = "humidity"    // component=sensor
	DeviceClassPressure    = "pressure"    // component=sensor
	DeviceClassBattery     = "battery"     // component=sensor
	DeviceClassIlluminance = "illuminance" // component=sensor
	DeviceClassShade       = "shade"       // component=cover
)

// keys: https://www.home-assistant.io/docs/mqtt/discovery/#configuration-variables
// many of the "omitempty" attributes are strictly required
type DiscoveryOptions struct {
	Name                string      `json:"name"`
	DeviceClass         string      `json:"device_class,omitempty"`  // https://www.home-assistant.io/integrations/binary_sensor/ | https://www.home-assistant.io/integrations/sensor/
	StateTopic          string      `json:"state_topic,omitempty"`   // required
	CommandTopic        string      `json:"command_topic,omitempty"` // optional (unset if isn't commandable)
	JsonAttributesTopic string      `json:"json_attributes_topic,omitempty"`
	ValueTemplate       string      `json:"value_template,omitempty"`
	PayloadOn           interface{} `json:"payload_on,omitempty"`  // can be a string describing MQTT message or a boolean indicating property value resolved by *ValueTemplate*
	PayloadOff          interface{} `json:"payload_off,omitempty"` // same as for *on*
	UniqueId            string      `json:"unique_id,omitempty"`
	Icon                string      `json:"icon,omitempty"` // e.g. "mdi:gesture-double-tap"
	UnitOfMeasurement   string      `json:"unit_of_measurement,omitempty"`

	Schema string `json:"schema,omitempty"` // use 'json' to receive commands in JSON. only applicable for controllable things, like lights (at least not applicable for sensors)

	Device DiscoveryOptionsDevice `json:"device"`

	Optimistic bool `json:"optimistic,omitempty"` // whether Home Assistant just assumes that the command it sent succeeded

	PositionOpen   *int `json:"position_open,omitempty"`
	PositionClosed *int `json:"position_closed,omitempty"`

	// capabilities, when using these you probably need schema=json

	Brightness bool `json:"brightness,omitempty"` // can control brightness
	ColorTemp  bool `json:"color_temp,omitempty"` // can control color temperature
	XY         bool `json:"xy,omitempty"`         // can control color
}

type DiscoveryOptionsDevice struct {
	Name            string   `json:"name,omitempty"`
	Manufacturer    string   `json:"manufacturer,omitempty"`
	Model           string   `json:"model,omitempty"`
	SoftwareVersion string   `json:"sw_version,omitempty"`
	AreaSuggested   string   `json:"suggested_area,omitempty"`
	Identifiers     []string `json:"identifiers,omitempty"`
}

type State struct {
	State      string            `json:"state"`
	Attributes map[string]string `json:"attributes"`
}

type Entity struct {
	Id        string
	name      string
	component Component

	discoveryOpts DiscoveryOptions
}

func NewSensor(id string, name string, deviceClass string, prefix TopicPrefix, hasAttributesTopic bool) *Entity {
	return &Entity{
		Id:        id,
		name:      name,
		component: ComponentSensor,

		discoveryOpts: DiscoveryOptions{
			Name:                id,
			DeviceClass:         deviceClass,
			StateTopic:          fmt.Sprintf("homeassistant/%s/hautomo/%s/state", component, id),
			JsonAttributesTopic: fmt.Sprintf("homeassistant/%s/hautomo/%s/attributes", component, id),
		},
	}
}

// https://www.home-assistant.io/integrations/switch.mqtt/
func NewSwitchEntity(id string, name string, opts DiscoveryOptions) *Entity {
	return NewEntityWithDiscoveryOpts(id, ComponentSwitch, name, opts)
}

// https://www.home-assistant.io/integrations/light.mqtt/
func NewLightEntity(id string, name string, opts DiscoveryOptions) *Entity {
	return NewEntityWithDiscoveryOpts(id, ComponentLight, name, opts)
}

// https://www.home-assistant.io/integrations/sensor.mqtt/
func NewSensorEntity(id string, name string, opts DiscoveryOptions) *Entity {
	return NewEntityWithDiscoveryOpts(id, ComponentSensor, name, opts)
}

// https://www.home-assistant.io/integrations/binary_sensor.mqtt
func NewBinarySensorEntity(id string, name string, opts DiscoveryOptions) *Entity {
	return NewEntityWithDiscoveryOpts(id, ComponentBinarySensor, name, opts)
}

// https://www.home-assistant.io/integrations/cover.mqtt
func NewCoverEntity(id string, name string, opts DiscoveryOptions) *Entity {
	return NewEntityWithDiscoveryOpts(id, ComponentCover, name, opts)
}

func NewEntityWithDiscoveryOpts(
	id string,
	component Component,
	name string,
	opts DiscoveryOptions) *Entity {
	// TODO: this is dirty
	if opts.Name == "" {
		opts.Name = name
	}

	return &Entity{
		Id:        id,
		name:      name,
		component: component,

		discoveryOpts: opts,
	}
}

func (h *Entity) mqttDiscoveryTopic() []byte {
	// to debug existing autodiscovery configs made by other programs, like zigbee2mqtt, run:
	//   $ docker exec -it mosquitto mosquitto_sub -t homeassistant/#

	// <discovery_prefix>/<component>/[<node_id>/]<object_id>/config
	return []byte(fmt.Sprintf("homeassistant/%s/hautomo/%s/config", h.component, h.Id))
}

func (h *Entity) mqttDiscoveryMsg() []byte {
	msg, err := json.Marshal(h.discoveryOpts)
	if err != nil {
		panic(err)
	}
	return msg
}
