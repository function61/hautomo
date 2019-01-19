package hapitypes

import (
	"fmt"
)

// for zigbee devices see https://koenkk.github.io/zigbee2mqtt/information/supported_devices.html
var deviceTypes = map[string]*DeviceType{
	"ikea-trådfri-noncolored": &DeviceType{
		Name:         "Trådfri non-colored",
		Manufacturer: "IKEA",
		Model:        "todo",
		Capabilities: Capabilities{
			Power:            true,
			Brightness:       true,
			ColorTemperature: true,
		},
	},
	"ikea-trådfri-rgb": &DeviceType{
		Name:         "Trådfri RGB",
		Manufacturer: "IKEA",
		Model:        "todo",
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
		Model:        "todo",
		Capabilities: Capabilities{
			Power: true,
		},
	},
	"ledstrip-rgb": &DeviceType{
		Name:         "LED strip RGB",
		Manufacturer: "Generic",
		Model:        "Generic",
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
		Capabilities: Capabilities{
			Power: true,
		},
	},
	"tv-philips-55PUS7909": &DeviceType{
		Name:         "Philips 55PUS7909",
		Manufacturer: "Philips",
		Model:        "55PUS7909",
		Capabilities: Capabilities{
			Power: true,
		},
	},
	"aqara-temperature-humidity": &DeviceType{
		Name:         "Aqara temperature/humidity sensor",
		Manufacturer: "Xiaomi",
		Model:        "WSDCGQ11LM",
		BatteryType:  "CR2032",
	},
	"aqara-water-leak": &DeviceType{
		Name:         "Aqara water leak sensor",
		Manufacturer: "Xiaomi",
		Model:        "SJCGQ11LM",
		BatteryType:  "CR2032",
	},
	"aqara-doorwindow": &DeviceType{
		Name:         "Aqara door & window contact sensor",
		Manufacturer: "Xiaomi",
		Model:        "MCCGQ11LM",
		BatteryType:  "CR1632",
	},
	"aqara-button": &DeviceType{
		Name:         "Aqara wireless button",
		Manufacturer: "Xiaomi",
		Model:        "WXKG11LM",
		BatteryType:  "CR2032",
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
	Capabilities Capabilities
}

type Capabilities struct {
	Power                     bool `json:"power"`
	Brightness                bool `json:"brightness"`
	Color                     bool `json:"color"`
	ColorTemperature          bool `json:"colortemperature"`
	ColorSeparateWhiteChannel bool `json:"color_separate_white_channel"`
	Playback                  bool `json:"playback"`
}
