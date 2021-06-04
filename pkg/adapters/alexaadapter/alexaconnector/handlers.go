// Implements various Alexa home automation APIs in AWS Lambda and serves as an
// anti-corruption layer hiding the misery (overcomplicated comms), sending only the
// relevant commands to Hautomo via SQS queue (so no direct connection required)
package alexaconnector

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/function61/gokit/crypto/cryptoutil"
	"github.com/function61/gokit/encoding/jsonfile"
	"github.com/function61/hautomo/pkg/adapters/alexaadapter/aamessages"
	"github.com/function61/hautomo/pkg/adapters/alexaadapter/alexadevicesync"
	"github.com/function61/hautomo/pkg/alexatypes"
	"github.com/function61/hautomo/pkg/ezstack/ezhub/hubtypes"
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
	if err := jsonfile.UnmarshalDisallowUnknownFields(discoveryFile, &spec); err != nil {
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
	if err := h.sendCommand(ctx, directiveMsg, aamessages.Message{
		DeviceId: directiveMsg.Directive.Endpoint.EndpointId,
		Attrs: hubtypes.Attributes{
			On: h.bld().Bool(true),
		},
	}); err != nil {
		return nil, err
	}

	propResponse := h.mkProperty(directiveMsg.Directive.Header.Namespace, "powerState", "ON")

	return h.propertyResponse("Response", directiveMsg, propResponse), nil
}

func (h *Handlers) AlexaPowerControllerTurnOff(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaPowerControllerTurnOff,
) (*alexatypes.AlexaResponse, error) {
	if err := h.sendCommand(ctx, directiveMsg, aamessages.Message{
		DeviceId: directiveMsg.Directive.Endpoint.EndpointId,
		Attrs: hubtypes.Attributes{
			On: h.bld().Bool(false),
		},
	}); err != nil {
		return nil, err
	}

	propResponse := h.mkProperty(directiveMsg.Directive.Header.Namespace, "powerState", "OFF")

	return h.propertyResponse("Response", directiveMsg, propResponse), nil
}

func (h *Handlers) AlexaBrightnessControllerSetBrightness(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaBrightnessControllerSetBrightness,
) (*alexatypes.AlexaResponse, error) {
	if err := h.sendCommand(ctx, directiveMsg, aamessages.Message{
		DeviceId: directiveMsg.Directive.Endpoint.EndpointId,
		Attrs: hubtypes.Attributes{
			Brightness: h.bld().Int(int64(payload.Brightness)),
		},
	}); err != nil {
		return nil, err
	}

	propResponse := h.mkProperty(directiveMsg.Directive.Header.Namespace, "brightness", payload.Brightness)

	return h.propertyResponse("Response", directiveMsg, propResponse), nil
}

func (h *Handlers) AlexaPercentageControllerSetPercentage(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaPercentageControllerSetPercentage,
) (*alexatypes.AlexaResponse, error) {
	// TODO: assuming PercentageController is for cover control. that may not be the case in the future.
	if err := h.sendCommand(ctx, directiveMsg, aamessages.Message{
		DeviceId: directiveMsg.Directive.Endpoint.EndpointId,
		Attrs: hubtypes.Attributes{
			ShadePosition: h.bld().Int(int64(payload.Percentage)),
		},
	}); err != nil {
		return nil, err
	}

	propResponse := h.mkProperty(directiveMsg.Directive.Header.Namespace, "percentage", payload.Percentage)

	return h.propertyResponse("Response", directiveMsg, propResponse), nil
}

func (h *Handlers) AlexaColorTemperatureControllerSetColorTemperature(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaColorTemperatureControllerSetColorTemperature,
) (*alexatypes.AlexaResponse, error) {
	// https://visualsproducer.wordpress.com/2020/11/29/mireds-versus-degrees-kelvin-for-colour-temperature/
	mireds := math.Round(1_000_000 / float64(payload.ColorTemperatureInKelvin))

	if err := h.sendCommand(ctx, directiveMsg, aamessages.Message{
		DeviceId: directiveMsg.Directive.Endpoint.EndpointId,
		Attrs: hubtypes.Attributes{
			ColorTemperature: h.bld().Int(int64(mireds)),
		},
	}); err != nil {
		return nil, err
	}

	propResponse := h.mkProperty(
		directiveMsg.Directive.Header.Namespace,
		"colorTemperatureInKelvin",
		payload.ColorTemperatureInKelvin)

	return h.propertyResponse("Response", directiveMsg, propResponse), nil
}

func (h *Handlers) AlexaPlaybackControllerPlay(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaPlaybackControllerPlay,
) (*alexatypes.AlexaResponse, error) {
	return h.playbackCommon(ctx, hubtypes.PlaybackControlPlay, directiveMsg)
}

func (h *Handlers) AlexaPlaybackControllerPause(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaPlaybackControllerPause,
) (*alexatypes.AlexaResponse, error) {
	return h.playbackCommon(ctx, hubtypes.PlaybackControlPause, directiveMsg)
}

func (h *Handlers) AlexaPlaybackControllerStop(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaPlaybackControllerStop,
) (*alexatypes.AlexaResponse, error) {
	return h.playbackCommon(ctx, hubtypes.PlaybackControlStop, directiveMsg)
}

func (h *Handlers) AlexaPlaybackControllerStartOver(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaPlaybackControllerStartOver,
) (*alexatypes.AlexaResponse, error) {
	return h.playbackCommon(ctx, hubtypes.PlaybackControlStartOver, directiveMsg)
}

func (h *Handlers) AlexaPlaybackControllerPrevious(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaPlaybackControllerPrevious,
) (*alexatypes.AlexaResponse, error) {
	return h.playbackCommon(ctx, hubtypes.PlaybackControlPrevious, directiveMsg)
}

func (h *Handlers) AlexaPlaybackControllerNext(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaPlaybackControllerNext,
) (*alexatypes.AlexaResponse, error) {
	return h.playbackCommon(ctx, hubtypes.PlaybackControlNext, directiveMsg)
}

func (h *Handlers) AlexaPlaybackControllerRewind(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaPlaybackControllerRewind,
) (*alexatypes.AlexaResponse, error) {
	return h.playbackCommon(ctx, hubtypes.PlaybackControlRewind, directiveMsg)
}

func (h *Handlers) AlexaPlaybackControllerFastForward(
	ctx context.Context,
	directiveMsg *alexatypes.DirectiveInput,
	payload *alexatypes.AlexaPlaybackControllerFastForward,
) (*alexatypes.AlexaResponse, error) {
	return h.playbackCommon(ctx, hubtypes.PlaybackControlFastForward, directiveMsg)
}

func (h *Handlers) playbackCommon(
	ctx context.Context,
	control hubtypes.PlaybackControl,
	directiveMsg *alexatypes.DirectiveInput,
) (*alexatypes.AlexaResponse, error) {
	if err := h.sendCommand(ctx, directiveMsg, aamessages.Message{
		DeviceId: directiveMsg.Directive.Endpoint.EndpointId,
		Attrs: hubtypes.Attributes{
			PlaybackControl: h.bld().PlaybackControl(control),
		},
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

	to255 := func(input float64) uint8 { // [0...1.0] -> [0...255]
		return uint8(input * 255.0)
	}

	if err := h.sendCommand(ctx, directiveMsg, aamessages.Message{
		DeviceId: directiveMsg.Directive.Endpoint.EndpointId,
		Attrs: hubtypes.Attributes{
			Color: h.bld().Color(to255(color.R), to255(color.G), to255(color.B)),
		},
	}); err != nil {
		return nil, err
	}

	return h.propertyResponse("Response", directiveMsg, h.mkProperty(
		directiveMsg.Directive.Header.Namespace,
		"color",
		payload.Color)), nil
}

func (h *Handlers) bld() hubtypes.AttrBuilder { // helper
	return hubtypes.NewAttrBuilder(h.now().UTC())
}
