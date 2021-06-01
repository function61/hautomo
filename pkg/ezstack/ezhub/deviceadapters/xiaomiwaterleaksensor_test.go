package deviceadapters

import (
	"testing"

	"github.com/function61/gokit/testing/assert"
	"github.com/function61/hautomo/pkg/ezstack"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
)

func TestWaterLeakHeartbeatMessage(t *testing.T) {
	assert.EqualString(t, afIncomingMessageToMqttMsg(t, aqaraWaterLeakSensor(), "0000000035a80101009800d6e1020000431c5f114f0a050042156c756d692e73656e736f725f776c65616b2e61713101ff42220121d10b03281e0421a8130521060006240000000000082104020a21000064100035a81d"), `
{"voltage":3025,"battery":100}`)
}

func TestWaterLeakLeakMessage(t *testing.T) {
	assert.EqualString(t, afIncomingMessageToMqttMsg(t, aqaraWaterLeakSensor(), "0000000535a801010083001a8a02000009195300010000ff000035a81d"), `
{"water_leak":true}`)

	assert.EqualString(t, afIncomingMessageToMqttMsg(t, aqaraWaterLeakSensor(), "0000000535a80101006b00e8d402000009195400000000ff000035a81d"), `
{"water_leak":false}`)
}

func aqaraWaterLeakSensor() *ezstack.Device {
	return &ezstack.Device{
		ManufacturerId: 4151,
		Model:          modelAqaraWaterLeakSensor,
		LogicalType:    2,
		MainPowered:    false,
		PowerSource:    ezstack.Battery,
		NetworkAddress: "0xa835",
		IEEEAddress:    "0x00158d00024bfb83",
		Endpoints: []*ezstack.Endpoint{
			{
				Id:             1,
				ProfileId:      260,
				DeviceId:       1026,
				DeviceVersion:  1,
				InClusterList:  []cluster.ClusterId{0, 3, 1},
				OutClusterList: []cluster.ClusterId{25},
			},
		},
	}
}
