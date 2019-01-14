package zigbee2mqttadapter

type deviceKind int

const (
	deviceKindUnknown    deviceKind = iota
	deviceKindWXKG11LM              // button
	deviceKindMCCGQ11LM             // door/window sensor
	deviceKindSJCGQ11LM             // water leak
	deviceKindWSDCGQ11LM            // temperature
)

// TODO: how to guarantee that these are kept in-sync?
var deviceTypeToZ2mType = map[string]deviceKind{
	"aqara-button":               deviceKindWXKG11LM,
	"aqara-doorwindow":           deviceKindMCCGQ11LM,
	"aqara-water-leak":           deviceKindSJCGQ11LM,
	"aqara-temperature-humidity": deviceKindWSDCGQ11LM,
}

// TODO: these have more additional fields than just battery+voltage

// {"battery":100,"voltage":3055,"linkquality":47,"click":"double"}
type WXKG11LM struct {
	Battery uint    `json:"battery"`
	Voltage uint    `json:"voltage"`
	Click   *string `json:"click"` // single/double/... (unset if heartbeat)
}

// {"battery":100,"voltage":3085,"linkquality":52,"contact":true}
type MCCGQ11LM struct {
	Battery uint `json:"battery"`
	Voltage uint `json:"voltage"`
	Contact bool `json:"contact"`
}

// {"water_leak":false,"linkquality":49,"battery":100,"voltage":3055}
type SJCGQ11LM struct {
	Battery   uint `json:"battery"`
	Voltage   uint `json:"voltage"`
	WaterLeak bool `json:"water_leak"`
}

// {"temperature":24.04,"linkquality":89,"humidity":25.91,"pressure":963,"battery":100,"voltage":3135}
type WSDCGQ11LM struct {
	Battery     uint    `json:"battery"`
	Voltage     uint    `json:"voltage"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Pressure    float64 `json:"pressure"`
}
