package zigbee2mqttadapter

import (
	"encoding/json"
	"github.com/function61/gokit/assert"
	"strings"
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
			input: `{"battery":100,"voltage":3055,"linkquality":47,"click":"single"}`,
			kind:  deviceKindWXKG11LM,
			output: `PushButtonEvent {"Device":"dummyId","Specifier":"single"}
LinkQualityEvent {"Device":"dummyId","LinkQuality":47}
BatteryStatusEvent {"Device":"dummyId","BatteryPct":100,"Voltage":3055}`,
		},
		{
			input: `{"battery":100,"voltage":3055,"linkquality":47,"click":"double"}`,
			kind:  deviceKindWXKG11LM,
			output: `PushButtonEvent {"Device":"dummyId","Specifier":"double"}
LinkQualityEvent {"Device":"dummyId","LinkQuality":47}
BatteryStatusEvent {"Device":"dummyId","BatteryPct":100,"Voltage":3055}`,
		},
		{
			input: `{"battery":100,"voltage":3055,"linkquality":47}`,
			kind:  deviceKindWXKG11LM,
			output: `LinkQualityEvent {"Device":"dummyId","LinkQuality":47}
BatteryStatusEvent {"Device":"dummyId","BatteryPct":100,"Voltage":3055}`,
		},
		{
			input: `{"contact":true,"linkquality":70}`,
			kind:  deviceKindMCCGQ11LM,
			output: `ContactEvent {"Device":"dummyId","Contact":true}
LinkQualityEvent {"Device":"dummyId","LinkQuality":70}
BatteryStatusEvent {"Device":"dummyId","BatteryPct":0,"Voltage":0}`,
		},
		{
			input: `{"contact":false,"linkquality":70}`,
			kind:  deviceKindMCCGQ11LM,
			output: `ContactEvent {"Device":"dummyId","Contact":false}
LinkQualityEvent {"Device":"dummyId","LinkQuality":70}
BatteryStatusEvent {"Device":"dummyId","BatteryPct":0,"Voltage":0}`,
		},
		{
			input: `{"angle_x":2,"angle_y":0,"angle_z":88,"angle_x_absolute":88,"angle_y_absolute":90,"linkquality":68,"battery":100,"voltage":3115,"action":"vibration"}`,
			kind:  deviceKindDJT11LM,
			output: `VibrationEvent {"Device":"dummyId"}
LinkQualityEvent {"Device":"dummyId","LinkQuality":68}
BatteryStatusEvent {"Device":"dummyId","BatteryPct":100,"Voltage":3115}`,
		},
		{
			input: `{"illuminance":60,"linkquality":68,"occupancy":true}`,
			kind:  deviceKindRTCGQ11LM,
			output: `MotionEvent {"Device":"dummyId","Movement":true,"Illuminance":60}
LinkQualityEvent {"Device":"dummyId","LinkQuality":68}`,
		},
		{
			input: `{"click":"right_double","linkquality":97}`,
			kind:  deviceKindWXKG02LM,
			output: `PushButtonEvent {"Device":"dummyId","Specifier":"right_double"}
LinkQualityEvent {"Device":"dummyId","LinkQuality":97}`,
		},
		{
			input: `{"action":"brightness_down_click","linkquality":34}`,
			kind:  deviceKindE1524,
			output: `PushButtonEvent {"Device":"dummyId","Specifier":"brightness_down_click"}
LinkQualityEvent {"Device":"dummyId","LinkQuality":34}`,
		},
		{
			input:  `{"this is": "unsupported payload type"}`,
			kind:   deviceKindUnknown,
			output: "unknown device kind for dummyId",
		},
	}

	for _, test := range tests {
		t.Run(test.output, func(t *testing.T) {
			events, err := parseMsgPayload(topic, func(_ string) *resolvedDevice {
				return &resolvedDevice{"dummyId", test.kind}
			}, test.input)

			if err != nil {
				assert.EqualString(t, err.Error(), test.output)
			} else {
				allSerialized := []string{}
				for _, event := range events {
					eventJson, err := json.Marshal(event)
					assert.Assert(t, err == nil)

					allSerialized = append(allSerialized, event.InboundEventType()+" "+string(eventJson))
				}

				assert.EqualString(t, strings.Join(allSerialized, "\n"), test.output)
			}
		})
	}
}
