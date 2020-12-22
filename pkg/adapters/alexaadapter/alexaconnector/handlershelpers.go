package alexaconnector

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/function61/hautomo/pkg/adapters/alexaadapter/aamessages"
	"github.com/function61/hautomo/pkg/alexatypes"
)

func (h *Handlers) Handle(ctx context.Context, input []byte) (*alexatypes.AlexaResponse, error) {
	directiveMsg, payloadGeneric, err := unmarshalDirective(input)
	if err != nil {
		return nil, err
	}

	switch payload := payloadGeneric.(type) {
	case *alexatypes.AlexaAuthorizationAcceptGrantInput:
		return h.AlexaAuthorizationAcceptGrantInput(ctx, directiveMsg, payload)
	case *alexatypes.AlexaDiscoveryDiscoverInput:
		return h.AlexaDiscoveryDiscoverInput(ctx, directiveMsg, payload)
	case *alexatypes.AlexaReportStateInput:
		return h.AlexaReportStateInput(ctx, directiveMsg, payload)
	case *alexatypes.AlexaPowerControllerTurnOn:
		return h.AlexaPowerControllerTurnOn(ctx, directiveMsg, payload)
	case *alexatypes.AlexaPowerControllerTurnOff:
		return h.AlexaPowerControllerTurnOff(ctx, directiveMsg, payload)
	case *alexatypes.AlexaBrightnessControllerSetBrightness:
		return h.AlexaBrightnessControllerSetBrightness(ctx, directiveMsg, payload)
	case *alexatypes.AlexaColorTemperatureControllerSetColorTemperature:
		return h.AlexaColorTemperatureControllerSetColorTemperature(ctx, directiveMsg, payload)
	case *alexatypes.AlexaColorControllerSetColor:
		return h.AlexaColorControllerSetColor(ctx, directiveMsg, payload)
	case *alexatypes.AlexaPlaybackControllerPlay:
		return h.AlexaPlaybackControllerPlay(ctx, directiveMsg, payload)
	case *alexatypes.AlexaPlaybackControllerPause:
		return h.AlexaPlaybackControllerPause(ctx, directiveMsg, payload)
	case *alexatypes.AlexaPlaybackControllerStop:
		return h.AlexaPlaybackControllerStop(ctx, directiveMsg, payload)
	case *alexatypes.AlexaPlaybackControllerStartOver:
		return h.AlexaPlaybackControllerStartOver(ctx, directiveMsg, payload)
	case *alexatypes.AlexaPlaybackControllerPrevious:
		return h.AlexaPlaybackControllerPrevious(ctx, directiveMsg, payload)
	case *alexatypes.AlexaPlaybackControllerNext:
		return h.AlexaPlaybackControllerNext(ctx, directiveMsg, payload)
	case *alexatypes.AlexaPlaybackControllerRewind:
		return h.AlexaPlaybackControllerRewind(ctx, directiveMsg, payload)
	case *alexatypes.AlexaPlaybackControllerFastForward:
		return h.AlexaPlaybackControllerFastForward(ctx, directiveMsg, payload)
	default:
		return nil, fmt.Errorf(
			"unknown payload type: %s/%s",
			directiveMsg.Directive.Header.Namespace,
			directiveMsg.Directive.Header.Name)
	}
}

func (h *Handlers) sendCommand(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	command aamessages.Message,
) error {
	queue, found := directiveMsg.Directive.Endpoint.Cookie["queue"]
	if !found {
		return fmt.Errorf(
			"sendCommand: queue not specified for %s",
			directiveMsg.Directive.Endpoint.EndpointId)
	}

	return h.extSystems.SendCommand(ctx, queue, command)
}

func (h *Handlers) mkProperty(
	namespace string,
	propertyName string,
	propertyValue interface{},
) alexatypes.AlexaProperty {
	return alexatypes.AlexaProperty{
		Namespace:                 namespace,
		Name:                      propertyName,
		Value:                     propertyValue,
		TimeOfSample:              h.now(),
		UncertaintyInMilliseconds: 0,
	}
}

func (h *Handlers) propertyResponse(
	name string,
	directiveMsg *alexatypes.DirectiveInput,
	properties ...alexatypes.AlexaProperty,
) *alexatypes.AlexaResponse {
	noPayload := json.RawMessage("{}")

	return &alexatypes.AlexaResponse{
		Context: &alexatypes.AlexaResponseContext{
			Properties: properties,
		},
		Event: alexatypes.AlexaEvent{
			Header: alexatypes.AlexaEventHeader{
				Namespace:        "Alexa",
				Name:             name,
				PayloadVersion:   "3",
				MessageId:        h.msgIdGenerator(),
				CorrelationToken: directiveMsg.Directive.Header.CorrelationToken,
			},
			Endpoint: directiveMsg.Directive.Endpoint,
			Payload:  &noPayload,
		},
	}
}

// no context, correlation token, endpoint or payload
func (h *Handlers) basicResponse(
	namespace string,
	name string,
) *alexatypes.AlexaResponse {
	noPayload := json.RawMessage("{}")

	return &alexatypes.AlexaResponse{
		Event: alexatypes.AlexaEvent{
			Header: alexatypes.AlexaEventHeader{
				Namespace:      namespace,
				Name:           name,
				PayloadVersion: "3",
				MessageId:      h.msgIdGenerator(),
			},
			Payload: &noPayload,
		},
	}
}
