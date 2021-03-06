package hapitypes

import (
	"fmt"
)

// for zigbee devices see https://koenkk.github.io/zigbee2mqtt/information/supported_devices.html
var deviceTypes = map[string]*DeviceType{
	"ikea-trådfri-noncolored": &DeviceType{
		Name:         "Trådfri non-colored E14",
		Manufacturer: "IKEA",
		Model:        "LED1536G5",
		Class:        DeviceClassLight,
		Capabilities: Capabilities{
			Power:            true,
			Brightness:       true,
			ColorTemperature: true,
		},
	},
	"ikea-trådfri-rgb": &DeviceType{
		Name:         "Trådfri RGB E27",
		Manufacturer: "IKEA",
		Model:        "LED1624G9",
		Class:        DeviceClassLight,
		Capabilities: Capabilities{
			Power:            true,
			Brightness:       true,
			Color:            true,
			ColorTemperature: true,
		},
	},
	"ikea-trådfri-smartplug": &DeviceType{
		Name:         "Trådfri smartplug",
		Manufacturer: "IKEA",
		Model:        "E1603",
		Class:        DeviceClassSmartPlug, // user is expected to override this with more specific one in device conf
		Capabilities: Capabilities{
			Power: true,
		},
	},
	"ikea-fyrtur": &DeviceType{
		Name:         "Fyrtur roller blind",
		Manufacturer: "IKEA",
		Model:        "Fyrtur",
		BatteryType:  "Custom battery pack",
		Class:        DeviceClassRollerBlind,
		Capabilities: Capabilities{
			CoverPosition: true,
		},
	},
	"ikea-trådfri-remote": &DeviceType{
		Name:         "Trådfri remote",
		Manufacturer: "IKEA",
		Model:        "E1524",
		Class:        DeviceClassRemote,
		BatteryType:  "CR2032",
	},
	"ledstrip-rgb": &DeviceType{
		Name:         "LED strip RGB",
		Manufacturer: "Generic",
		Model:        "Generic",
		Class:        DeviceClassLight,
		Capabilities: Capabilities{
			Power:      true,
			Brightness: true,
			Color:      true,
		},
	},
	"ledstrip-rgbw": &DeviceType{
		Name:         "LED strip RGB(W)",
		Manufacturer: "Generic",
		Model:        "Generic",
		Class:        DeviceClassLight,
		Capabilities: Capabilities{
			Power:                     true,
			Brightness:                true,
			Color:                     true,
			ColorSeparateWhiteChannel: true,
		},
	},
	"onkyo-tx-nr515": &DeviceType{
		Name:         "Onkyo TX-NR515",
		Manufacturer: "Onkyo",
		Model:        "TX-NR515",
		Class:        DeviceClassAmplifier,
		Capabilities: Capabilities{
			Power: true,
		},
	},
	"sonoff-basic": &DeviceType{
		Name:         "Sonoff Basic",
		Manufacturer: "Sonoff",
		Model:        "Sonoff Basic",
		Class:        DeviceClassSmartPlug, // user is expected to override this with more specific one in device conf
		Capabilities: Capabilities{
			Power: true,
		},
	},
	"tv-philips-55PUS7909": &DeviceType{
		Name:         "Philips 55PUS7909",
		Manufacturer: "Philips",
		Model:        "55PUS7909",
		Class:        DeviceClassTV,
		Capabilities: Capabilities{
			Power: true,
		},
	},
	"aqara-temperature-humidity": &DeviceType{
		Name:         "Aqara temperature/humidity sensor",
		Manufacturer: "Xiaomi",
		Model:        "WSDCGQ11LM",
		BatteryType:  "CR2032",
		Class:        DeviceClassClimateSensor,
		Capabilities: Capabilities{
			ReportsTemperature: true,
		},
	},
	"aqara-water-leak": &DeviceType{
		Name:         "Aqara water leak sensor",
		Manufacturer: "Xiaomi",
		Model:        "SJCGQ11LM",
		BatteryType:  "CR2032",
		Class:        DeviceClassSensor,
	},
	"aqara-motion-sensor": &DeviceType{
		Name:         "Aqara motion sensor",
		Manufacturer: "Xiaomi",
		Model:        "RTCGQ11LM",
		BatteryType:  "CR2450",
		Class:        DeviceClassPresenceSensor,
	},
	"aqara-doorwindow": &DeviceType{
		Name:         "Aqara door & window contact sensor",
		Manufacturer: "Xiaomi",
		Model:        "MCCGQ11LM",
		BatteryType:  "CR1632",
		Class:        DeviceClassDoor, // might also be window sensor, but defaulting to this more common use case
	},
	"aqara-vibration-sensor": &DeviceType{
		Name:         "Aqara vibration sensor",
		Manufacturer: "Xiaomi",
		Model:        "DJT11LM",
		BatteryType:  "CR2032",
		Class:        DeviceClassSensor, // user is expected to specify exact type of sensor
	},
	"aqara-button": &DeviceType{
		Name:         "Aqara wireless button",
		Manufacturer: "Xiaomi",
		Model:        "WXKG11LM",
		BatteryType:  "CR2032",
		Class:        DeviceClassRemote,
	},
	"aqara-doublekeyswitch": &DeviceType{
		Name:         "Aqara wireless double key switch",
		Manufacturer: "Xiaomi",
		Model:        "WXKG02LM",
		BatteryType:  "CR2032",
		Class:        DeviceClassRemote,
	},
	"eventghostClient": &DeviceType{
		Name:         "EventGhost client",
		Manufacturer: "EventGhost",
		Model:        "EventGhost",
		Class:        DeviceClassComputer,
		Capabilities: Capabilities{
			Playback: true,
		},
	},
	"screen-server:screen": &DeviceType{
		Name:         "Screen-server screen",
		Manufacturer: "function61.com",
		Class:        DeviceClassDisplay,
	},
	"virtual-switch": &DeviceType{
		Name:         "Virtual switch",
		Manufacturer: "function61.com",
		Class:        DeviceClassGeneric, // user is expected to override this in device conf
		Capabilities: Capabilities{
			Power:         true,
			VirtualSwitch: true,
		},
	},
}

func ResolveDeviceType(t string) (*DeviceType, error) {
	typ, found := deviceTypes[t]
	if !found {
		return nil, fmt.Errorf("device type not found: %s", t)
	}

	return typ, nil
}

type DeviceType struct {
	Name         string
	Manufacturer string
	Model        string
	BatteryType  string
	LinkToManual string
	Class        *DeviceClass // broad categorization of the device - its "icon"
	Capabilities Capabilities
}

type Capabilities struct {
	Power                     bool `json:"power"`
	Brightness                bool `json:"brightness"`
	Color                     bool `json:"color"`
	ColorTemperature          bool `json:"colortemperature"`
	ColorSeparateWhiteChannel bool `json:"color_separate_white_channel"`
	Playback                  bool `json:"playback"`
	ReportsTemperature        bool `json:"reports_temperature"`
	VirtualSwitch             bool `json:"virtual_switch"` // can send fake contact sensor triggers to Alexa to trigger routines
	CoverPosition             bool `json:"cover_position"`
}
