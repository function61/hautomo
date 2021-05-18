// Implements various Alexa home automation APIs in AWS Lambda and serves as an
// anti-corruption layer hiding the misery (overcomplicated comms), sending only the
// relevant commands to Hautomo via SQS queue (so no direct connection required)
package alexaconnector

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/function61/gokit/crypto/cryptoutil"
	"github.com/function61/gokit/encoding/jsonfile"
	"github.com/function61/hautomo/pkg/adapters/alexaadapter/aamessages"
	"github.com/function61/hautomo/pkg/adapters/alexaadapter/alexadevicesync"
	"github.com/function61/hautomo/pkg/alexatypes"
	"github.com/lucasb-eyer/go-colorful"
)

type messageIdGenerator func() string

type timeGetter func() time.Time

type HandlerOutput struct {
	Queue    string
	QueueMsg aamessages.Message
	Response *alexatypes.AlexaResponse
}

type Handlers struct {
	msgIdGenerator messageIdGenerator
	now            timeGetter
	extSystems     ExternalSystems
}

func New(extSystems ExternalSystems, msgIdGenerator messageIdGenerator, now timeGetter) *Handlers {
	if msgIdGenerator == nil {
		msgIdGenerator = func() string {
			return cryptoutil.RandHex(8)
		}
	}

	if now == nil {
		now = time.Now
	}

	return &Handlers{msgIdGenerator, now, extSystems}
}

func (h *Handlers) AlexaAuthorizationAcceptGrantInput(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaAuthorizationAcceptGrantInput,
) (*alexatypes.AlexaResponse, error) {
	// we don't actually do anything with this currently, so just log it if we need it for debug
	log.Printf("AcceptGrant %s", payload.Grant.Code)

	return h.basicResponse(directiveMsg.Directive.Header.Namespace, "AcceptGrant.Response"), nil
}

func (h *Handlers) AlexaDiscoveryDiscoverInput(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaDiscoveryDiscoverInput,
) (*alexatypes.AlexaResponse, error) {
	/*	We have user's access token in the discovery payload

		1) Resolve access token to user's (Amazon) User ID

		2) Fetch discovery file based on UID

		3) Translate discovery file into Alexa's JSON definitions
	*/
	if payload.Scope.Type != "BearerToken" {
		return nil, fmt.Errorf("unsupported token type: %s", payload.Scope.Type)
	}

	// 1)

	userId, err := h.extSystems.TokenToUserId(ctx, payload.Scope.Token)
	if err != nil {
		return nil, fmt.Errorf("TokenToUserId: %w", err)
	}

	// 2)

	discoveryFile, err := h.extSystems.FetchDiscoveryFile(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("FetchDiscoveryFile: %w", err)
	}
	defer discoveryFile.Close()

	// 3)

	spec := alexadevicesync.AlexaConnectorSpec{}
	if err := jsonfile.Unmarshal(discoveryFile, &spec, true); err != nil {
		return nil, err
	}

	resp, err := discoveryFileToAlexaInterfaces(spec, h.msgIdGenerator)
	if err != nil {
		return nil, fmt.Errorf("discoveryFileToAlexaInterfaces: %w", err)
	}

	return resp, nil
}

func (h *Handlers) AlexaReportStateInput(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaReportStateInput,
) (*alexatypes.AlexaResponse, error) {
	// TODO: add this as an actual type? (duplicated anyway)
	connectivityOk := struct {
		Value string `json:"value"`
	}{
		Value: "OK",
	}

	return h.propertyResponse(
		"StateReport",
		directiveMsg,
		h.mkProperty("Alexa.ContactSensor", "detectionState", "NOT_DETECTED"),
		h.mkProperty("Alexa.EndpointHealth", "connectivity", connectivityOk)), nil
}

func (h *Handlers) AlexaPowerControllerTurnOn(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaPowerControllerTurnOn,
) (*alexatypes.AlexaResponse, error) {
	powerOn := h.mkProperty(directiveMsg.Directive.Header.Namespace, "powerState", "ON")

	if err := h.sendCommand(ctx, directiveMsg, &aamessages.TurnOnRequest{
		DeviceIdOrDeviceGroupId: directiveMsg.Directive.Endpoint.EndpointId,
	}); err != nil {
		return nil, err
	}

	return h.propertyResponse("Response", directiveMsg, powerOn), nil
}

func (h *Handlers) AlexaPowerControllerTurnOff(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaPowerControllerTurnOff,
) (*alexatypes.AlexaResponse, error) {
	powerOff := h.mkProperty(directiveMsg.Directive.Header.Namespace, "powerState", "OFF")

	if err := h.sendCommand(ctx, directiveMsg, &aamessages.TurnOffRequest{
		DeviceIdOrDeviceGroupId: directiveMsg.Directive.Endpoint.EndpointId,
	}); err != nil {
		return nil, err
	}

	return h.propertyResponse("Response", directiveMsg, powerOff), nil
}

func (h *Handlers) AlexaBrightnessControllerSetBrightness(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaBrightnessControllerSetBrightness,
) (*alexatypes.AlexaResponse, error) {
	brightness := h.mkProperty(directiveMsg.Directive.Header.Namespace, "brightness", payload.Brightness)

	if err := h.sendCommand(ctx, directiveMsg, &aamessages.BrightnessRequest{
		DeviceIdOrDeviceGroupId: directiveMsg.Directive.Endpoint.EndpointId,
		Brightness:              uint(payload.Brightness),
	}); err != nil {
		return nil, err
	}

	return h.propertyResponse("Response", directiveMsg, brightness), nil
}

func (h *Handlers) AlexaColorTemperatureControllerSetColorTemperature(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaColorTemperatureControllerSetColorTemperature,
) (*alexatypes.AlexaResponse, error) {
	colorTemperatureInKelvin := h.mkProperty(
		directiveMsg.Directive.Header.Namespace,
		"colorTemperatureInKelvin",
		payload.ColorTemperatureInKelvin)

	if err := h.sendCommand(ctx, directiveMsg, &aamessages.ColorTemperatureRequest{
		DeviceIdOrDeviceGroupId:  directiveMsg.Directive.Endpoint.EndpointId,
		ColorTemperatureInKelvin: uint(payload.ColorTemperatureInKelvin),
	}); err != nil {
		return nil, err
	}

	return h.propertyResponse("Response", directiveMsg, colorTemperatureInKelvin), nil
}

func (h *Handlers) AlexaPlaybackControllerPlay(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaPlaybackControllerPlay,
) (*alexatypes.AlexaResponse, error) {
	return h.playbackCommon(ctx, "Play", directiveMsg)
}

func (h *Handlers) AlexaPlaybackControllerPause(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaPlaybackControllerPause,
) (*alexatypes.AlexaResponse, error) {
	return h.playbackCommon(ctx, "Pause", directiveMsg)
}

func (h *Handlers) AlexaPlaybackControllerStop(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaPlaybackControllerStop,
) (*alexatypes.AlexaResponse, error) {
	return h.playbackCommon(ctx, "Stop", directiveMsg)
}

func (h *Handlers) AlexaPlaybackControllerStartOver(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaPlaybackControllerStartOver,
) (*alexatypes.AlexaResponse, error) {
	return h.playbackCommon(ctx, "StartOver", directiveMsg)
}

func (h *Handlers) AlexaPlaybackControllerPrevious(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaPlaybackControllerPrevious,
) (*alexatypes.AlexaResponse, error) {
	return h.playbackCommon(ctx, "Previous", directiveMsg)
}

func (h *Handlers) AlexaPlaybackControllerNext(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaPlaybackControllerNext,
) (*alexatypes.AlexaResponse, error) {
	return h.playbackCommon(ctx, "Next", directiveMsg)
}

func (h *Handlers) AlexaPlaybackControllerRewind(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaPlaybackControllerRewind,
) (*alexatypes.AlexaResponse, error) {
	return h.playbackCommon(ctx, "Rewind", directiveMsg)
}

func (h *Handlers) AlexaPlaybackControllerFastForward(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaPlaybackControllerFastForward,
) (*alexatypes.AlexaResponse, error) {
	return h.playbackCommon(ctx, "FastForward", directiveMsg)
}

func (h *Handlers) playbackCommon(
	ctx context.Context,
	action string,
	directiveMsg *alexatypes.DirectiveInput,
) (*alexatypes.AlexaResponse, error) {
	if err := h.sendCommand(ctx, directiveMsg, &aamessages.PlaybackRequest{
		DeviceIdOrDeviceGroupId: directiveMsg.Directive.Endpoint.EndpointId,
		Action:                  action,
	}); err != nil {
		return nil, err
	}

	return h.propertyResponse("Response", directiveMsg), nil
}

func (h *Handlers) AlexaColorControllerSetColor(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaColorControllerSetColor,
) (*alexatypes.AlexaResponse, error) {
	color := colorful.Hsv(
		payload.Color.Hue,
		payload.Color.Saturation,
		payload.Color.Brightness)

	if err := h.sendCommand(ctx, directiveMsg, &aamessages.ColorRequest{
		DeviceIdOrDeviceGroupId: directiveMsg.Directive.Endpoint.EndpointId,
		Red:                     uint8(color.R * 255.0),
		Green:                   uint8(color.G * 255.0),
		Blue:                    uint8(color.B * 255.0),
	}); err != nil {
		return nil, err
	}

	return h.propertyResponse("Response", directiveMsg, h.mkProperty(
		directiveMsg.Directive.Header.Namespace,
		"color",
		payload.Color)), nil
}
