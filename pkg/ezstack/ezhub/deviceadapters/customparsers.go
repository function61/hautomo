package deviceadapters

import (
	"github.com/function61/hautomo/pkg/ezstack"
)

// we have to list devices here only for devices that we've to have custom parsers for
// TODO: nuke this list, as it's mainly needed os we don't have to define repeat the model ID for test code
const (
	modelAqaraButtonSensor       ezstack.Model = "lumi.sensor_switch.aq2"            // https://www.zigbee2mqtt.io/devices/WXKG11LM.html
	modelAqaraDoubleButtonSensor ezstack.Model = "lumi.remote.b286acn01\x00\x00\x00" // triple-null terminated, how else...
	modelAqaraPresenceSensor     ezstack.Model = "lumi.sensor_motion.aq2"            // https://www.zigbee2mqtt.io/devices/RTCGQ01LM.html
	modelAqaraTemperatureSensor  ezstack.Model = "lumi.weather"
	modelAqaraWaterLeakSensor    ezstack.Model = "lumi.sensor_wleak.aq1"
	modelAqaraVibrationSensor    ezstack.Model = "lumi.vibration.aq1"
	modelAqaraDoorSensor         ezstack.Model = "lumi.sensor_magnet.aq2"
	modelIkeaRollerBlind         ezstack.Model = "FYRTUR block-out roller blind"
)
