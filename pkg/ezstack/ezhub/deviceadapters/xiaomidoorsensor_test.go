package deviceadapters

import (
	"testing"

	"github.com/function61/gokit/testing/assert"
	"github.com/function61/hautomo/pkg/ezstack"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
)

func TestDoorSensorHeartbeatMessage(t *testing.T) {
	assert.EqualString(t, afIncomingMessageToMqttMsg(t, aqaraDoorSensor(), "0000000088d50101005e00dc4c010000261c5f11010a01ff421d0121b30b03281f0421a80105210201062401000000000a21000064100188d51d"), `
{"voltage":2995,"battery":97}`)
}

func TestDoorSensorOpen(t *testing.T) {
	assert.EqualString(t, afIncomingMessageToMqttMsg(t, aqaraDoorSensor(), "0000060088d50101006e0016b44200000718030a0000100188d51d"), `
{"contact":true}`)
}

func TestDoorSensorClose(t *testing.T) {
	assert.EqualString(t, afIncomingMessageToMqttMsg(t, aqaraDoorSensor(), "0000060088d50101006e005c024500000718040a0000100088d51d"), `
{"contact":false}`)
}

func aqaraDoorSensor() *ezstack.Device {
	return &ezstack.Device{
		Manufacturer:   "LUMI",
		ManufacturerId: 4151,
		Model:          modelAqaraDoorSensor,
		LogicalType:    2,
		MainPowered:    false,
		PowerSource:    ezstack.Battery,
		NetworkAddress: "0xd588",
		IEEEAddress:    "0x00158d0002b52dce",
		Endpoints: []*ezstack.Endpoint{
			{
				Id:             1,
				ProfileId:      260,
				DeviceId:       24321,
				DeviceVersion:  1,
				InClusterList:  []cluster.ClusterId{0, 3, 65535, 6},
				OutClusterList: []cluster.ClusterId{0, 4, 65535},
			},
		},
	}
}
