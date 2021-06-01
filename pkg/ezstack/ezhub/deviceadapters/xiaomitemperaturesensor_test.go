package deviceadapters

import (
	"testing"

	"github.com/function61/gokit/testing/assert"
	"github.com/function61/hautomo/pkg/ezstack"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
)

func TestTemperatureSensor(t *testing.T) {
	// JFC the firmware sends three different Zigbee frames for three different measurements,
	// wasting energy for the device (and others in the network as well). this seems to be a limitation
	// of Zigbee as temperature/humidity/pressure are all different clusters and while one
	// AttributeReport message can send many different readings, they all must concern the same cluster..

	assert.EqualString(t, afIncomingMessageToMqttMsg(t, aqaraTemperatureSensor(), "00000204ba0201010088009999fd00000818e70a000029ce09ba021d"), `
{"temperature":25.1}`)

	assert.EqualString(t, afIncomingMessageToMqttMsg(t, aqaraTemperatureSensor(), "00000504ba020101008600a699fd00000818e80a000021f919ba021d"), `
{"humidity":66.49}`)

	assert.EqualString(t, afIncomingMessageToMqttMsg(t, aqaraTemperatureSensor(), "00000304ba020101008600b499fd00001118e90a000029df03140028ff100029b626ba021d"), `
{"pressure":991}`)
}

func TestTemperatureSensorHeartbeatMessage(t *testing.T) {
	assert.EqualString(t, afIncomingMessageToMqttMsg(t, aqaraTemperatureSensor(), "00000000ba02010100560080b60000002c18de0a01ff42250121d10b0421a8130521b7ab0624000000000064291b0a65216810662bfd8201000a210000ba021d"), `
{"voltage":3025,"battery":100}`)
}

func aqaraTemperatureSensor() *ezstack.Device {
	return &ezstack.Device{
		Manufacturer:   "LUMI",
		ManufacturerId: 4151,
		Model:          modelAqaraTemperatureSensor,
		LogicalType:    2,
		MainPowered:    false,
		PowerSource:    ezstack.Battery,
		NetworkAddress: "0x02ba",
		IEEEAddress:    "0x00158d000272c6bf",
		Endpoints: []*ezstack.Endpoint{
			{
				Id:             1,
				ProfileId:      260,
				DeviceId:       24321,
				DeviceVersion:  1,
				InClusterList:  []cluster.ClusterId{0, 3, 65535, 1026, 1027, 1029},
				OutClusterList: []cluster.ClusterId{0, 4, 65535},
			},
		},
	}
}
