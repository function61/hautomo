package deviceadapters

import (
	"testing"

	"github.com/function61/gokit/testing/assert"
	"github.com/function61/hautomo/pkg/ezstack"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
)

func TestPresenceSensorHeartbeatMessage(t *testing.T) {
	assert.EqualString(t, afIncomingMessageToMqttMsg(t, aqaraPresenceSensor(), "00000000f1fe0101009300854a9b00002a1c5f11870a01ff42210121e50b0328190421a81305214700062403000000000a2100006410000b215602f1fe1d"), `
{"voltage":3045,"battery":100}`)
}

func TestPresenceSensor(t *testing.T) {
	// it sends illuminance and occupancy in different messages

	assert.EqualString(t, afIncomingMessageToMqttMsg(t, aqaraPresenceSensor(), "00000004f1fe010100930062eae400000818880a0000218900f1fe1d"), `
{"illuminance":137}`)

	assert.EqualString(t, afIncomingMessageToMqttMsg(t, aqaraPresenceSensor(), "00000604f1fe010100930070eae400000718890a00001801f1fe1d"), `
{"occupancy":true}`)
}

func aqaraPresenceSensor() *ezstack.Device {
	return &ezstack.Device{
		Manufacturer:   "LUMI",
		ManufacturerId: 4151,
		Model:          modelAqaraPresenceSensor,
		LogicalType:    2,
		MainPowered:    false,
		PowerSource:    ezstack.Battery,
		NetworkAddress: "0xfef1",
		IEEEAddress:    "0x00158d0003021058",
		Endpoints: []*ezstack.Endpoint{
			{
				Id:             1,
				ProfileId:      260,
				DeviceVersion:  1,
				InClusterList:  []cluster.ClusterId{0, 65535, 1030, 1024, 1280, 1, 3},
				OutClusterList: []cluster.ClusterId{0, 25},
			},
		},
	}
}
