package deviceadapters

import (
	"testing"

	"github.com/function61/gokit/testing/assert"
	"github.com/function61/hautomo/pkg/evdevcodes"
	"github.com/function61/hautomo/pkg/ezstack"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
)

func TestIKEAE1524Remote(t *testing.T) {
	// center button
	_, attrs := afIncomingMessageToAttributes(t, ikeaE1524Remote(), "05210600547b01010054007cbc0300000301390212bc0a")
	assert.Assert(t, attrs.Press.Key == evdevcodes.KeyPOWER)

	// brightness up
	_, attrs = afIncomingMessageToAttributes(t, ikeaE1524Remote(), "05210800547b01010054005cd503000007013a06002b050012bc0a")
	assert.Assert(t, attrs.Press.Key == evdevcodes.KeyBRIGHTNESSUP)

	// brightness down
	_, attrs = afIncomingMessageToAttributes(t, ikeaE1524Remote(), "05210800547b010100560013f103000007013b02012b050012bc0a")
	assert.Assert(t, attrs.Press.Key == evdevcodes.KeyBRIGHTNESSDOWN)

	// left
	_, attrs = afIncomingMessageToAttributes(t, ikeaE1524Remote(), "05210500547b0101005400de0c04000009057c113c0701010d0012bc0a")
	assert.Assert(t, attrs.Press.Key == evdevcodes.KeyLEFT)

	// right
	_, attrs = afIncomingMessageToAttributes(t, ikeaE1524Remote(), "05210500547b0101005600e21f04000009057c113d0700010d0012bc0a")
	assert.Assert(t, attrs.Press.Key == evdevcodes.KeyRIGHT)
}

func ikeaE1524Remote() *ezstack.Device {
	return &ezstack.Device{
		Manufacturer:   "IKEA of Sweden",
		ManufacturerId: 4476,
		Model:          "TRADFRI remote control",
		LogicalType:    2,
		MainPowered:    false,
		PowerSource:    ezstack.Battery,
		NetworkAddress: "0x7b54",
		IEEEAddress:    "0x90fd9ffffee8cdc2",
		Endpoints: []*ezstack.Endpoint{
			{
				Id:             1,
				ProfileId:      49246,
				DeviceId:       2096,
				DeviceVersion:  2,
				InClusterList:  []cluster.ClusterId{0, 1, 3, 9, 2821, 4096},
				OutClusterList: []cluster.ClusterId{3, 4, 5, 6, 8, 25, 4096},
			},
		},
	}
}
