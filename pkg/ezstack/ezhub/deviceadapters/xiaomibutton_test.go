package deviceadapters

import (
	"testing"

	"github.com/function61/gokit/testing/assert"
	"github.com/function61/hautomo/pkg/ezstack"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
)

func TestButtonSensorHeartbeatMessage(t *testing.T) {
	assert.EqualString(t, afIncomingMessageToMqttMsg(t, aqaraButtonSensor(), "00000000bc820101008d009fe9bc00003d1c5f11000a050042166c756d692e73656e736f725f7377697463682e61713201ff421a0121c70b0328210421a81305214b00062400000000000a210000bc821d"), `
{"voltage":3015,"battery":100}`)
}

func TestButtonSensorSingleClick(t *testing.T) {
	assert.EqualString(t, afIncomingMessageToMqttMsg(t, aqaraButtonSensor(), "00000600bc820101006e0063aec900000b18010a0000100000001001bc821d"), `
{"action":"single"}`)
}

func TestButtonSensorDoubleClick(t *testing.T) {
	assert.EqualString(t, afIncomingMessageToMqttMsg(t, aqaraButtonSensor(), "00000600bc820101007b00fb81cb00000718020a00802002bc821d"), `
{"action":"double"}`)
}

func TestButtonSensorTripleClick(t *testing.T) {
	assert.EqualString(t, afIncomingMessageToMqttMsg(t, aqaraButtonSensor(), "00000600bc820101006b003f43cd00000718030a00802003bc821d"), `
{"action":"triple"}`)

	// for some reason it also sends additional message (but only when >= 3 clicks)

	assert.EqualString(t, afIncomingMessageToMqttMsg(t, aqaraButtonSensor(), "00000000bc820101006b004c43cd0000111c5f11040a01ff42090421a8130a210000bc821d"), `
{}`)
}

func TestButtonSensorQuadrupleClick(t *testing.T) {
	assert.EqualString(t, afIncomingMessageToMqttMsg(t, aqaraButtonSensor(), "00000600bc820101007e009855d500000718070a00802004bc821d"), `
{"action":"quadruple"}`)

	// for some reason it also sends additional message (but only when >= 3 clicks)

	assert.EqualString(t, afIncomingMessageToMqttMsg(t, aqaraButtonSensor(), "00000000bc820101007b00a655d50000111c5f11080a01ff42090421a8130a210000bc821d"), `
{}`)
}

func aqaraButtonSensor() *ezstack.Device {
	return &ezstack.Device{
		Manufacturer:   "LUMI",
		ManufacturerId: 4151,
		Model:          modelAqaraButtonSensor,
		LogicalType:    2,
		MainPowered:    false,
		PowerSource:    ezstack.Battery,
		NetworkAddress: "0x82bc",
		IEEEAddress:    "0x00158d000204a208",
		Endpoints: []*ezstack.Endpoint{
			{
				Id:             1,
				ProfileId:      260,
				DeviceId:       24321,
				DeviceVersion:  1,
				InClusterList:  []cluster.ClusterId{0, 65535, 6},
				OutClusterList: []cluster.ClusterId{0, 4, 65535},
			},
		},
	}
}
