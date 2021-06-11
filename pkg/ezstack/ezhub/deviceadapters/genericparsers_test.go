package deviceadapters

import (
	"testing"

	"github.com/function61/gokit/testing/assert"
	"github.com/function61/hautomo/pkg/ezstack"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
)

func TestIkeaRollerBlind(t *testing.T) {
	_, attrs := afIncomingMessageToAttributes(t, ikeaRollerBlind(), "00000201933d0101007b000d9a8a00000708440a08002000ade21c")

	assert.Assert(t, attrs.ShadePosition.Value == 0)

	_, attrs = afIncomingMessageToAttributes(t, ikeaRollerBlind(), "00000201933d0101007b00c1c63d000007086f0a08002024ade21c")

	assert.Assert(t, attrs.ShadePosition.Value == 36)
}

func ikeaRollerBlind() *ezstack.Device {
	return &ezstack.Device{
		Manufacturer:   "IKEA of Sweden",
		ManufacturerId: 4476,
		Model:          modelIkeaRollerBlind,
		LogicalType:    2,
		MainPowered:    false,
		PowerSource:    ezstack.Battery,
		NetworkAddress: "0x3d93",
		IEEEAddress:    "0x842e14fffe6a5cda",
		Endpoints: []*ezstack.Endpoint{
			{
				Id:             1,
				ProfileId:      260,
				DeviceId:       514,
				DeviceVersion:  1,
				InClusterList:  []cluster.ClusterId{0, 1, 3, 4, 5, 32, 258, 4096},
				OutClusterList: []cluster.ClusterId{25, 4096},
			},
		},
	}
}
