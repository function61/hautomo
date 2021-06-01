package deviceadapters

import (
	"testing"

	"github.com/function61/gokit/testing/assert"
	"github.com/function61/hautomo/pkg/evdevcodes"
	"github.com/function61/hautomo/pkg/ezstack"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
)

func TestShadeRemote(t *testing.T) {
	_, attrs := afIncomingMessageToAttributes(t, ikeaTradfriOpenCloseRemote(), "00000201e0f20101001700c8bf00000003010500e0f20b")
	assert.Assert(t, attrs.Press.Key == evdevcodes.KeyOPEN)

	_, attrs = afIncomingMessageToAttributes(t, ikeaTradfriOpenCloseRemote(), "00000201e0f2010100150022d600000003010601e0f20b")
	assert.Assert(t, attrs.Press.Key == evdevcodes.KeyCLOSE)
}

func ikeaTradfriOpenCloseRemote() *ezstack.Device {
	return &ezstack.Device{
		Manufacturer:   "\x02KE",
		ManufacturerId: 4476,
		Model:          "TRADFRI open/close remote",
		LogicalType:    2,
		MainPowered:    false,
		PowerSource:    ezstack.Battery,
		NetworkAddress: "0xf2e0",
		IEEEAddress:    "0x842e14fffe667d3b",
		Endpoints: []*ezstack.Endpoint{
			{
				Id:             1,
				ProfileId:      260,
				DeviceId:       24321,
				DeviceVersion:  1,
				InClusterList:  []cluster.ClusterId{0, 1, 3, 9, 32, 4096, 64636},
				OutClusterList: []cluster.ClusterId{3, 4, 6, 8, 25, 258, 4096},
			},
		},
	}
}
