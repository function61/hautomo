package hapitypes

import (
	"github.com/function61/hautomo/pkg/homeassistant"
)

// device's main type category. e.g. light, without going deeper into its characteristics
// (RGB light, color temperature controllable). you can think of this as device type icon ("light bulb").
type DeviceClass struct {
	FriendlyName      string // user-friendly name
	HomeAssistantIcon homeassistant.IconId
	AlexaCategory     string // available values https://developer.amazon.com/docs/device-apis/alexa-discovery.html#display-categories
}

// if AlexaCategory is OTHER, it means it probably cannot be represented in Alexa

var (
	DeviceClassGeneric        = &DeviceClass{"Generic", homeassistant.IconEye, "OTHER"}
	DeviceClassLight          = &DeviceClass{"Light", homeassistant.IconLightbulb, "LIGHT"}
	DeviceClassRollerBlind    = &DeviceClass{"Roller blind", homeassistant.IconBlinds, "INTERIOR_BLIND"}
	DeviceClassRemote         = &DeviceClass{"Remote", homeassistant.IconRemote, "OTHER"}
	DeviceClassAmplifier      = &DeviceClass{"Amplifier", homeassistant.IconSpeaker, "SPEAKER"}
	DeviceClassTV             = &DeviceClass{"TV", homeassistant.IconTelevision, "TV"}
	DeviceClassClimateSensor  = &DeviceClass{"Climate sensor", homeassistant.IconThermometer, "TEMPERATURE_SENSOR"}
	DeviceClassPresenceSensor = &DeviceClass{"Presence sensor", homeassistant.IconHome, "MOTION_SENSOR"}
	DeviceClassSensor         = &DeviceClass{"Sensor", homeassistant.IconGauge, "OTHER"}
	DeviceClassFan            = &DeviceClass{"Fan", homeassistant.IconFan, "FAN"}
	DeviceClassDoor           = &DeviceClass{"Door", homeassistant.IconDoorClosed, "DOOR"}
	DeviceClassComputer       = &DeviceClass{"Computer", homeassistant.IconTablet, "COMPUTER"}
	DeviceClassDisplay        = &DeviceClass{"Display", homeassistant.IconMonitor, "SCREEN"}
	DeviceClassSmartPlug      = &DeviceClass{"Smart plug", homeassistant.IconPower, "SMARTPLUG"} // used only if user doesn't specify a specific device connected to the smart plug
	DeviceClassSleepingSensor = &DeviceClass{"Sleeping sensor", homeassistant.IconSleep, "WEARABLE"}
)

var DeviceClassById = map[string]*DeviceClass{
	"Generic":        DeviceClassGeneric,
	"Light":          DeviceClassLight,
	"RollerBlind":    DeviceClassRollerBlind,
	"Remote":         DeviceClassRemote,
	"Amplifier":      DeviceClassAmplifier,
	"TV":             DeviceClassTV,
	"ClimateSensor":  DeviceClassClimateSensor,
	"PresenceSensor": DeviceClassPresenceSensor,
	"Sensor":         DeviceClassSensor,
	"Fan":            DeviceClassFan,
	"Door":           DeviceClassDoor,
	"Computer":       DeviceClassComputer,
	"Display":        DeviceClassDisplay,
	"SmartPlug":      DeviceClassSmartPlug,
	"SleepingSensor": DeviceClassSleepingSensor,
}
