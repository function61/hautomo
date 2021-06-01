package deviceadapters

import (
	"testing"

	"github.com/function61/gokit/testing/assert"
	"github.com/function61/hautomo/pkg/ezstack"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
)

func TestVibrationSensor(t *testing.T) {
	assert.EqualString(t, afIncomingMessageToMqttMsg(t, aqaraVibrationSensor(), "0000010184a5010100880029fcf600000a18ed0a0505230000160084a51d"), `
{"extra":{"strength":"22"}}`)

	assert.EqualString(t, afIncomingMessageToMqttMsg(t, aqaraVibrationSensor(), "0000010184a501010083001f096000000d18f40a5500210200030521100084a51d"), `
{"action":"tilt","extra":{"angle":"16"}}`)

	assert.EqualString(t, afIncomingMessageToMqttMsg(t, aqaraVibrationSensor(), "0000010184a50101008b004a366100000a18f60a05052300001d0084a51d"), `
{"extra":{"strength":"29"}}`)

	assert.EqualString(t, afIncomingMessageToMqttMsg(t, aqaraVibrationSensor(), "0000010184a50101008600452d6000000c18f50a0805252700ffff910484a51d"), `
{"angle_x":0,"angle_y":89,"angle_z":1}`)
}

func TestVibrationSensorHeartbeatMessage(t *testing.T) {
	assert.EqualString(t, afIncomingMessageToMqttMsg(t, aqaraVibrationSensor(), "0000000084a50101007e00ff56320000371c5f11e00a01ff422e0121e50b03281b0421a8130521760006240700000000082108030a21000098211400992153009a2529000000920484a51d"), `
{"voltage":3045,"battery":100}`)
}

func aqaraVibrationSensor() *ezstack.Device {
	return &ezstack.Device{
		ManufacturerId: 4151,
		Model:          modelAqaraVibrationSensor,
		LogicalType:    2,
		MainPowered:    false,
		PowerSource:    ezstack.Battery,
		NetworkAddress: "0xa584",
		IEEEAddress:    "0x00158d0002b12519",
		Endpoints: []*ezstack.Endpoint{
			{
				Id:             1,
				ProfileId:      260,
				DeviceId:       10,
				DeviceVersion:  1,
				InClusterList:  []cluster.ClusterId{0, 3, 25, 257},
				OutClusterList: []cluster.ClusterId{0, 4, 3, 5, 25, 257},
			},
		},
	}
}
