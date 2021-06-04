package alexaconnector

import (
	"encoding/json"
	"fmt"

	"github.com/function61/hautomo/pkg/alexatypes"
)

var namespaceAndNameAllocators = map[string]func() interface{}{
	"Alexa.Authorization#AcceptGrant":                      func() interface{} { return &alexatypes.AlexaAuthorizationAcceptGrantInput{} },
	"Alexa.Discovery#Discover":                             func() interface{} { return &alexatypes.AlexaDiscoveryDiscoverInput{} },
	"Alexa#ReportState":                                    func() interface{} { return &alexatypes.AlexaReportStateInput{} },
	"Alexa.PowerController#TurnOn":                         func() interface{} { return &alexatypes.AlexaPowerControllerTurnOn{} },
	"Alexa.PowerController#TurnOff":                        func() interface{} { return &alexatypes.AlexaPowerControllerTurnOff{} },
	"Alexa.BrightnessController#SetBrightness":             func() interface{} { return &alexatypes.AlexaBrightnessControllerSetBrightness{} },
	"Alexa.ColorTemperatureController#SetColorTemperature": func() interface{} { return &alexatypes.AlexaColorTemperatureControllerSetColorTemperature{} },
	"Alexa.ColorController#SetColor":                       func() interface{} { return &alexatypes.AlexaColorControllerSetColor{} },
	"Alexa.PlaybackController#Play":                        func() interface{} { return &alexatypes.AlexaPlaybackControllerPlay{} },
	"Alexa.PlaybackController#Pause":                       func() interface{} { return &alexatypes.AlexaPlaybackControllerPause{} },
	"Alexa.PlaybackController#Stop":                        func() interface{} { return &alexatypes.AlexaPlaybackControllerStop{} },
	"Alexa.PlaybackController#StartOver":                   func() interface{} { return &alexatypes.AlexaPlaybackControllerStartOver{} },
	"Alexa.PlaybackController#Previous":                    func() interface{} { return &alexatypes.AlexaPlaybackControllerPrevious{} },
	"Alexa.PlaybackController#Next":                        func() interface{} { return &alexatypes.AlexaPlaybackControllerNext{} },
	"Alexa.PlaybackController#Rewind":                      func() interface{} { return &alexatypes.AlexaPlaybackControllerRewind{} },
	"Alexa.PlaybackController#FastForward":                 func() interface{} { return &alexatypes.AlexaPlaybackControllerFastForward{} },
	"Alexa.PercentageController#SetPercentage":             func() interface{} { return &alexatypes.AlexaPercentageControllerSetPercentage{} },
}

func unmarshalDirective(inputJson []byte) (*alexatypes.DirectiveInput, interface{}, error) {
	directiveMsg := &alexatypes.DirectiveInput{}
	if err := json.Unmarshal(inputJson, directiveMsg); err != nil {
		return nil, nil, err
	}

	namespaceAndName := fmt.Sprintf(
		"%s#%s",
		directiveMsg.Directive.Header.Namespace,
		directiveMsg.Directive.Header.Name)

	allocator, found := namespaceAndNameAllocators[namespaceAndName]
	if !found {
		return nil, nil, fmt.Errorf("unsupported directive: %s", namespaceAndName)
	}

	payload := allocator()

	if err := json.Unmarshal(*directiveMsg.Directive.Payload, payload); err != nil {
		return nil, nil, fmt.Errorf("unmarshal failed for %s: %w", namespaceAndName, err)
	}

	return directiveMsg, payload, nil
}
