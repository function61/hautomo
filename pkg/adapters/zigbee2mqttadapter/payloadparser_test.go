package zigbee2mqttadapter

import (
	"encoding/json"
	"github.com/function61/gokit/assert"
	"testing"
)

func TestParseMsgPayload(t *testing.T) {
	// dummy
	topic := "zigbee2mqtt/0x00158d000227a73c"

	tests := []struct {
		input  string
		kind   deviceKind
		output string
	}{
		{
			input:  `{"battery":100,"voltage":3055,"linkquality":47,"click":"single"}`,
			kind:   deviceKindWXKG11LM,
			output: `PushButtonEvent {"Device":"dummyId","Specifier":"single"}`,
		},
		{
			input:  `{"battery":100,"voltage":3055,"linkquality":47,"click":"double"}`,
			kind:   deviceKindWXKG11LM,
			output: `PushButtonEvent {"Device":"dummyId","Specifier":"double"}`,
		},
		{
			input:  `{"battery":100,"voltage":3055,"linkquality":47}`,
			kind:   deviceKindWXKG11LM,
			output: `HeartbeatEvent {"Device":"dummyId"}`,
		},
		{
			input:  `{"contact":true,"linkquality":70}`,
			kind:   deviceKindMCCGQ11LM,
			output: `ContactEvent {"Device":"dummyId","Contact":true}`,
		},
		{
			input:  `{"contact":false,"linkquality":70}`,
			kind:   deviceKindMCCGQ11LM,
			output: `ContactEvent {"Device":"dummyId","Contact":false}`,
		},
		{
			input:  `{"this is": "unsupported payload type"}`,
			kind:   deviceKindUnknown,
			output: "unknown device kind for dummyId",
		},
	}

	for _, test := range tests {
		t.Run(test.output, func(t *testing.T) {
			e, err := parseMsgPayload(topic, func(_ string) *resolvedDevice {
				return &resolvedDevice{"dummyId", test.kind}
			}, test.input)

			if err != nil {
				assert.EqualString(t, err.Error(), test.output)
			} else {
				eventJson, err := json.Marshal(e)
				assert.Assert(t, err == nil)
				assert.EqualString(t, e.InboundEventType()+" "+string(eventJson), test.output)
			}
		})
	}
}
