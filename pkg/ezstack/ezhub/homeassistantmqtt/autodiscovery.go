package homeassistantmqtt

import (
	"fmt"

	"github.com/function61/hautomo/pkg/ezstack"
	"github.com/function61/hautomo/pkg/ezstack/ezhub/hubtypes"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
	"github.com/function61/hautomo/pkg/homeassistant"
)

func AutodiscoveryEntities(dev *hubtypes.Device, mqttPrefix string) []*homeassistant.Entity {
	entities := []*homeassistant.Entity{}

	if dev.Area == "" { // skip devices without area specified
		return entities
	}

	addEntity := func(entity *homeassistant.Entity) { // helper
		entities = append(entities, entity)
	}

	id := dev.ZigbeeDevice.IEEEAddress

	uniqueId := func(entityType string) string {
		return fmt.Sprintf("%s_%s_hautomo", id, entityType)
	}

	devSpec := homeassistant.DiscoveryOptionsDevice{
		Name:         dev.FriendlyName,
		Manufacturer: dev.ZigbeeDevice.Manufacturer,
		Model:        string(dev.ZigbeeDevice.Model),

		Identifiers: []string{id},

		AreaSuggested: dev.Area,

		SoftwareVersion: "Hautomo-EZhub",
	}

	stateTopic := fmt.Sprintf("%s/%s", mqttPrefix, id)
	commandTopic := fmt.Sprintf("%s/%s/set", mqttPrefix, id)

	// the "primary entities" don't have name suffix
	// (i.e. contact sensor is just "<friendly name>" and not "<friendly name> - contact")

	// all Zigbee devices have this entity
	addEntity(homeassistant.NewSensorEntity(
		id+"_linkquality",
		dev.FriendlyName+" - link quality",
		homeassistant.DiscoveryOptions{
			UniqueId: uniqueId("linkquality"),

			StateTopic: stateTopic,

			ValueTemplate:     "{{ value_json.linkquality }}",
			UnitOfMeasurement: "-",

			Icon: "mdi:radio-tower",

			Device: devSpec,
		}))

	// FIXME: is it safe assumption that all level-controllable things are lights?
	if dev.ImplementsCluster(cluster.IdGenLevelCtrl) {
		addEntity(homeassistant.NewLightEntity(
			id+"_light",
			dev.FriendlyName,
			homeassistant.DiscoveryOptions{
				UniqueId: uniqueId("light"),

				StateTopic:   stateTopic,
				CommandTopic: commandTopic,

				Schema: "json",

				Brightness: dev.ImplementsCluster(cluster.IdGenLevelCtrl),
				ColorTemp:  dev.ImplementsCluster(cluster.IdLightingColorCtrl),
				XY:         dev.ImplementsCluster(cluster.IdLightingColorCtrl),

				Device: devSpec,
			}))
	}

	if dev.ImplementsCluster(cluster.IdMsTemperatureMeasurement) {
		addEntity(homeassistant.NewSensorEntity(
			id+"_temp",
			dev.FriendlyName+" - temp",
			homeassistant.DiscoveryOptions{
				DeviceClass: homeassistant.DeviceClassTemperature,
				UniqueId:    uniqueId("temperature"),

				StateTopic: stateTopic,

				ValueTemplate:     "{{ value_json.temperature }}",
				UnitOfMeasurement: "°C",

				Device: devSpec,
			}))
	}

	if dev.ImplementsCluster(cluster.IdMsRelativeHumidity) {
		addEntity(homeassistant.NewSensorEntity(
			id+"_humid",
			dev.FriendlyName+" - humidity",
			homeassistant.DiscoveryOptions{
				DeviceClass: homeassistant.DeviceClassHumidity,
				UniqueId:    uniqueId("humidity"),

				StateTopic: stateTopic,

				ValueTemplate:     "{{ value_json.humidity }}",
				UnitOfMeasurement: "%",

				Device: devSpec,
			}))
	}

	if dev.ImplementsCluster(cluster.IdMsPressureMeasurement) {
		addEntity(homeassistant.NewSensorEntity(
			id+"_pressure",
			dev.FriendlyName+" - pressure",
			homeassistant.DiscoveryOptions{
				DeviceClass: homeassistant.DeviceClassPressure,
				UniqueId:    uniqueId("pressure"),

				StateTopic: stateTopic,

				ValueTemplate:     "{{ value_json.pressure }}",
				UnitOfMeasurement: "hPa",

				Device: devSpec,
			}))
	}

	if dev.ImplementsCluster(cluster.IdClosuresWindowCovering) {
		addEntity(homeassistant.NewCoverEntity(
			id+"_shade",
			dev.FriendlyName,
			homeassistant.DiscoveryOptions{
				DeviceClass: homeassistant.DeviceClassShade,
				UniqueId:    uniqueId("shade"),

				StateTopic:   stateTopic,
				CommandTopic: commandTopic,

				Optimistic: true,

				Device: devSpec,
			}))
	}

	// FIXME: being battery powered does not necessarily mean we get the voltage reported to us
	if dev.ZigbeeDevice.PowerSource == ezstack.Battery {
		addEntity(homeassistant.NewSensorEntity(
			id+"_battery",
			dev.FriendlyName+" - battery",
			homeassistant.DiscoveryOptions{
				DeviceClass: homeassistant.DeviceClassBattery,
				UniqueId:    uniqueId("battery"),

				StateTopic: stateTopic,

				ValueTemplate:     "{{ value_json.battery }}",
				UnitOfMeasurement: "%",

				Device: devSpec,
			}))
	}

	switch dev.ZigbeeDevice.Model {
	case "lumi.sensor_switch.aq2":
		// cannot be binary sensor, because we need to be able to deal with double clicks etc.
		// TODO: docs recommend: https://www.home-assistant.io/integrations/device_trigger.mqtt/
		addEntity(homeassistant.NewSensorEntity(
			id+"_click",
			dev.FriendlyName+" - click",
			homeassistant.DiscoveryOptions{
				UniqueId: uniqueId("click"),

				StateTopic: stateTopic,
				// PayloadOn:    `{"state": "ON"}`,
				// PayloadOff:   `{"state": "OFF"}`,
				// PayloadOn:    "ON",
				// PayloadOff:   "OFF",
				ValueTemplate: "{{ value_json.action }}",

				Icon: "mdi:toggle-switch",

				Device: devSpec,
			}))
	case "lumi.sensor_magnet.aq2":
		addEntity(homeassistant.NewBinarySensorEntity(
			id+"_contact",
			dev.FriendlyName,
			homeassistant.DiscoveryOptions{
				DeviceClass: homeassistant.DeviceClassDoor,
				UniqueId:    uniqueId("contact"),

				StateTopic: stateTopic,

				ValueTemplate: "{{ value_json.contact }}",
				PayloadOn:     true,
				PayloadOff:    false,

				Device: devSpec,
			}))
	case "lumi.sensor_motion.aq2":
		addEntity(homeassistant.NewBinarySensorEntity(
			id+"_occupancy",
			dev.FriendlyName,
			homeassistant.DiscoveryOptions{
				DeviceClass: homeassistant.DeviceClassOccupancy,
				UniqueId:    uniqueId("occupancy"),

				StateTopic: stateTopic,

				ValueTemplate: "{{ value_json.occupancy }}",
				PayloadOn:     true,
				PayloadOff:    false,

				Device: devSpec,
			}))

		addEntity(homeassistant.NewSensorEntity(
			id+"_illuminance",
			dev.FriendlyName+" - illuminance",
			homeassistant.DiscoveryOptions{
				DeviceClass: homeassistant.DeviceClassIlluminance,
				UniqueId:    uniqueId("illuminance"),

				StateTopic: stateTopic,

				ValueTemplate:     "{{ value_json.illuminance }}",
				UnitOfMeasurement: "lx", // TODO: verify?

				Device: devSpec,
			}))
	}

	return entities
}
