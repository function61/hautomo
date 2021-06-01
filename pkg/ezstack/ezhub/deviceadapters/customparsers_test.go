package deviceadapters

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/function61/gokit/testing/assert"
	"github.com/function61/hautomo/pkg/ezstack"
	"github.com/function61/hautomo/pkg/ezstack/binstruct"
	"github.com/function61/hautomo/pkg/ezstack/ezhub/homeassistantmqtt"
	"github.com/function61/hautomo/pkg/ezstack/ezhub/hubtypes"
	"github.com/function61/hautomo/pkg/ezstack/zcl"
	"github.com/function61/hautomo/pkg/ezstack/zigbee"
	"github.com/function61/hautomo/pkg/ezstack/znp"
	"github.com/function61/hautomo/pkg/ezstack/znp/unp"
)

var (
	staticTimestamp = time.Date(2021, 5, 16, 11, 27, 0, 0, time.UTC)
)

// parses msg like CommandType=2 Subsystem=4 Command=129 Payload=...
func afIncomingMessageToMqttMsg(t *testing.T, dev *ezstack.Device, payloadHex string) string {
	t.Helper()

	incomingMessage, attrs := afIncomingMessageToAttributes(t, dev, payloadHex)

	batteryType := For(dev).BatteryType()

	wdev := &hubtypes.Device{
		ZigbeeDevice: dev,

		State: &hubtypes.DeviceState{
			// LinkQuality: &hubtypes.AttrInt{Value: 0, LastReport: staticTimestamp},

			EndpointAttrs: map[zigbee.EndpointId]*hubtypes.Attributes{
				incomingMessage.SrcEndpoint: attrs,
			},
		},
	}

	// wdev.State.LinkQuality = nil

	msg, err := homeassistantmqtt.MessageFromChangedAttributes(
		attrs,
		wdev,
		batteryType,
		staticTimestamp)
	assert.Ok(t, err)

	// go through the zigbee2mqtt layer for easy serialization
	return "\n" + msg
}

func afIncomingMessageToAttributes(t *testing.T, dev *ezstack.Device, payloadHex string) (*zcl.ZclIncomingMessage, *hubtypes.Attributes) {
	t.Helper()

	payload, err := hex.DecodeString(payloadHex)
	assert.Ok(t, err)

	command, err := znp.NewConcreteAsyncCommand(unp.S_AF, znp.AfIncomingMessageId)
	assert.Ok(t, err)

	// .. which we'll fill dynamically here
	assert.Ok(t, binstruct.Decode(payload, command))

	incomingMessage, err := zcl.Library.ToZclIncomingMessage(command.(*znp.AfIncomingMessage))
	assert.Ok(t, err)

	assert.EqualString(t, incomingMessage.SrcAddr, dev.NetworkAddress)

	actx := &hubtypes.AttrsCtx{hubtypes.NewAttributes(), staticTimestamp, incomingMessage.SrcEndpoint}

	assert.Ok(t, ZclIncomingMessageToAttributes(incomingMessage, actx, dev))

	return incomingMessage, actx.Attrs
}

/*
func afIncomingMessageToAttributesJson(t *testing.T, dev *ezstack.Device, payloadHex string) string {
	t.Helper()

	_, attrs := afIncomingMessageToAttributes(t, dev, payloadHex)

	jb, err := json.MarshalIndent(attrs, "", "  ")
	assert.Ok(t, err)

	return string(jb)
}
*/
