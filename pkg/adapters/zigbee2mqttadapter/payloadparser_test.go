package zigbee2mqttadapter

import (
	"github.com/function61/gokit/assert"
	"testing"
)

func TestParseMsgPayload(t *testing.T) {
	// dummy
	topic := "zigbee2mqtt/0x00158d000227a73c"

	tests := []struct {
		input  string
		output string
	}{
		{
			input:  `{"battery":100,"voltage":3055,"linkquality":47,"click":"single"}`,
			output: "zigbee2mqtt:0x00158d000227a73c:click",
		},
		{
			input:  `{"battery":100,"voltage":3055,"linkquality":47,"click":"double"}`,
			output: "zigbee2mqtt:0x00158d000227a73c:double",
		},
		{
			input:  `{"contact":true,"linkquality":70}`,
			output: "zigbee2mqtt:0x00158d000227a73c:contact:true",
		},
		{
			input:  `{"contact":false,"linkquality":70}`,
			output: "zigbee2mqtt:0x00158d000227a73c:contact:false",
		},
		{
			input:  `{"this is": "unsupported payload type"}`,
			output: "",
		},
	}

	for _, test := range tests {
		t.Run(test.output, func(t *testing.T) {
			e := parseMsgPayload(topic, test.input)

			if test.output != "" {
				assert.EqualString(t, e.Event, test.output)
			} else {
				assert.True(t, e == nil)
			}
		})
	}
}
