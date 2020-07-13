package alexadevicesync

import (
	"encoding/json"
	"testing"

	"github.com/function61/gokit/assert"
	"github.com/function61/hautomo/pkg/hapitypes"
)

func TestCreateAlexaConnectorSpec(t *testing.T) {
	conf := &hapitypes.ConfigFile{
		Adapters: []hapitypes.AdapterConfig{
			{
				SqsQueueUrl:           "http://dummy.com/queue",
				SqsAlexaUsertokenHash: "usertokenhash",
			},
		},
		Devices: []hapitypes.DeviceConfig{
			{
				DeviceId:      "dev1",
				Name:          "Kitchen light",
				AlexaCategory: "SMARTPLUG",
				Type:          "onkyo-tx-nr515",
			},
			{
				DeviceId:      "dev2",
				Name:          "Balcony light",
				Description:   "RGBW light",
				AlexaCategory: "LIGHT",
				Type:          "ledstrip-rgbw",
			},
		},
	}

	spec, err := createAlexaConnectorSpec(conf.Adapters[0], conf)
	assert.Assert(t, err == nil)

	jsonBytes, err := json.MarshalIndent(spec, "", "  ")
	assert.Assert(t, err == nil)

	assert.EqualString(t, string(jsonBytes), `{
  "queue": "http://dummy.com/queue",
  "devices": [
    {
      "id": "dev1",
      "friendly_name": "Kitchen light",
      "description": "",
      "display_category": "SMARTPLUG",
      "capability_codes": [
        "PowerController"
      ]
    },
    {
      "id": "dev2",
      "friendly_name": "Balcony light",
      "description": "RGBW light",
      "display_category": "LIGHT",
      "capability_codes": [
        "PowerController",
        "BrightnessController",
        "ColorController"
      ]
    }
  ]
}`)
}
