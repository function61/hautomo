package alexadevicesync

import (
	"encoding/json"
	"github.com/function61/gokit/assert"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
	"testing"
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
				DeviceId:        "dev1",
				Name:            "Kitchen light",
				AlexaCategory:   "SMARTPLUG",
				CapabilityPower: true,
			},
			{
				DeviceId:                   "dev2",
				Name:                       "God",
				Description:                "Has all the capabilities",
				AlexaCategory:              "LIGHT",
				CapabilityPower:            true,
				CapabilityBrightness:       true,
				CapabilityColor:            true,
				CapabilityColorTemperature: true,
				CapabilityPlayback:         true,
			},
		},
	}

	spec, err := createAlexaConnectorSpec(conf.Adapters[0], conf)
	assert.True(t, err == nil)

	jsonBytes, err := json.MarshalIndent(spec, "", "  ")
	assert.True(t, err == nil)

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
      "friendly_name": "God",
      "description": "Has all the capabilities",
      "display_category": "LIGHT",
      "capability_codes": [
        "PowerController",
        "BrightnessController",
        "ColorController",
        "ColorTemperatureController",
        "PlaybackController"
      ]
    }
  ]
}`)
}
