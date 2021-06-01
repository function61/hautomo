package deviceadapters

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/function61/gokit/testing/assert"
	"github.com/function61/hautomo/pkg/ezstack"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
)

func TestXiaomiDoubleButton(t *testing.T) {
	for _, tc := range []struct {
		subject       string
		zigbeeMessage string
		expectedPress string
		expectedMqtt  string
	}{
		{
			"left",
			"00001200d0660101002a005cb90000000818020a5500210100d0661d",
			`{"key":256}`,
			"\n" + `{"action":"single_left"}`,
		},
		{
			"right",
			"00001200d066020100440058c70000000818030a5500210100d0661d",
			`{"key":257}`,
			"\n" + `{"action":"single_right"}`,
		},
		{
			"right double click",
			"00001200d066020100410093ee0000000818040a5500210200d0661d",
			`{"key":257,"count":2}`,
			"\n" + `{"action":"double_right"}`,
		},
		{
			"hold left & right",
			"00001200d066030100460014390100000818060a5500210000d0661d",
			`{"key":256,"keys_additional":[257],"kind":2}`,
			"\n" + `{"action":"hold_both"}`,
		},
	} {
		tc := tc // pin

		t.Run(tc.subject, func(t *testing.T) {
			_, attrs := afIncomingMessageToAttributes(t, xiaomiDoubleButtonSensor(), tc.zigbeeMessage)
			// assert.Assert(t, attrs.Press.Key == evdevcodes.Btn0)

			serialized, err := json.Marshal(attrs.Press)
			assert.Ok(t, err)

			assert.EqualString(t, strings.ReplaceAll(string(serialized), `,"reported":"2021-05-16T11:27:00Z"`, ""), tc.expectedPress)

			assert.EqualString(t, afIncomingMessageToMqttMsg(t, xiaomiDoubleButtonSensor(), tc.zigbeeMessage), tc.expectedMqtt)

		})
	}
}

func xiaomiDoubleButtonSensor() *ezstack.Device {
	return &ezstack.Device{
		Manufacturer:   "LUMI",
		ManufacturerId: 4151,
		Model:          modelAqaraDoubleButtonSensor,
		LogicalType:    2,
		MainPowered:    false,
		PowerSource:    ezstack.Battery,
		NetworkAddress: "0x66d0",
		IEEEAddress:    "0x00158d0002cb1a49",
		Endpoints: []*ezstack.Endpoint{
			{
				Id:             1,
				ProfileId:      260,
				DeviceId:       24321,
				DeviceVersion:  1,
				InClusterList:  []cluster.ClusterId{0, 3, 25, 65535, 18},
				OutClusterList: []cluster.ClusterId{0, 4, 3, 5, 25, 65535, 18},
			},
			{
				Id:             2,
				ProfileId:      260,
				DeviceId:       24322,
				DeviceVersion:  1,
				InClusterList:  []cluster.ClusterId{3, 18},
				OutClusterList: []cluster.ClusterId{4, 3, 5, 18},
			},
			{
				Id:             3,
				ProfileId:      260,
				DeviceId:       24323,
				DeviceVersion:  1,
				InClusterList:  []cluster.ClusterId{3, 12},
				OutClusterList: []cluster.ClusterId{4, 3, 5, 12},
			},
		},
	}
}
