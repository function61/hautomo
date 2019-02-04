package zigbee2mqttadapter

type deviceKind int

const (
	deviceKindUnknown    deviceKind = iota
	deviceKindWXKG11LM              // button
	deviceKindWXKG02LM              // double button switch
	deviceKindMCCGQ11LM             // door/window sensor
	deviceKindSJCGQ11LM             // water leak
	deviceKindWSDCGQ11LM            // temperature
	deviceKindRTCGQ11LM             // motion sensor
)

// TODO: how to guarantee that these are kept in-sync?
var deviceTypeToZ2mType = map[string]deviceKind{
	"aqara-button":               deviceKindWXKG11LM,
	"aqara-doublekeyswitch":      deviceKindWXKG02LM,
	"aqara-doorwindow":           deviceKindMCCGQ11LM,
	"aqara-water-leak":           deviceKindSJCGQ11LM,
	"aqara-temperature-humidity": deviceKindWSDCGQ11LM,
	"aqara-motion-sensor":        deviceKindRTCGQ11LM,
}

// {"battery":100,"voltage":3055,"linkquality":47,"click":"double"}
type WXKG11LM struct {
	Click       *string `json:"click"` // single/double/... (unset if heartbeat)
	Battery     uint    `json:"battery"`
	Voltage     uint    `json:"voltage"`
	LinkQuality uint    `json:"linkquality"`
}

// {"battery":100,"voltage":3085,"linkquality":52,"contact":true}
type MCCGQ11LM struct {
	Contact     bool `json:"contact"`
	Battery     uint `json:"battery"`
	Voltage     uint `json:"voltage"`
	LinkQuality uint `json:"linkquality"`
}

// {"water_leak":false,"linkquality":49,"battery":100,"voltage":3055}
type SJCGQ11LM struct {
	WaterLeak   bool `json:"water_leak"`
	Battery     uint `json:"battery"`
	Voltage     uint `json:"voltage"`
	LinkQuality uint `json:"linkquality"`
}

// {"illuminance":60,"linkquality":68,"occupancy":true}
type RTCGQ11LM struct {
	Occupancy   bool `json:"occupancy"`
	Illuminance uint `json:"illuminance"`
	LinkQuality uint `json:"linkquality"`
}

// {"temperature":24.04,"linkquality":89,"humidity":25.91,"pressure":963,"battery":100,"voltage":3135}
type WSDCGQ11LM struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Pressure    float64 `json:"pressure"`
	Battery     uint    `json:"battery"`
	Voltage     uint    `json:"voltage"`
	LinkQuality uint    `json:"linkquality"`
}

// {"click":"left","linkquality":97}
type WXKG02LM struct {
	Click       *string `json:"click"` // (probably unset if heartbeat) left|left_long|right|right_long|left_double|right_double|both|both_double
	LinkQuality uint    `json:"linkquality"`
}
