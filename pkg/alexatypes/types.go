// The missing types I couldn't find from anywhere in AWS's SDK or the Go ecosystem
package alexatypes

import (
	"encoding/json"
	"time"
)

type NoPayload struct{}

type TypeAndToken struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

type Endpoint struct {
	Scope      *TypeAndToken     `json:"scope,omitempty"`
	EndpointId string            `json:"endpointId"`
	Cookie     map[string]string `json:"cookie,omitempty"`
}

// shared in directive input & responses
type AlexaEventHeader struct {
	Namespace        string  `json:"namespace"`
	Name             string  `json:"name"`
	PayloadVersion   string  `json:"payloadVersion"`
	MessageId        string  `json:"messageId"`
	CorrelationToken *string `json:"correlationToken,omitempty"`
}

type AlexaEvent struct {
	Header   AlexaEventHeader `json:"header"`
	Endpoint *Endpoint        `json:"endpoint,omitempty"`
	Payload  *json.RawMessage `json:"payload"`
}

type DirectiveInput struct {
	Directive AlexaEvent `json:"directive"`
}

type AlexaProperty struct {
	Namespace                 string      `json:"namespace"`
	Name                      string      `json:"name"`
	Value                     interface{} `json:"value"` // can contain basically whatever..
	TimeOfSample              time.Time   `json:"timeOfSample"`
	UncertaintyInMilliseconds int         `json:"uncertaintyInMilliseconds"`
}

type AlexaResponseContext struct {
	Properties []AlexaProperty `json:"properties"`
}

type AlexaResponse struct {
	Context *AlexaResponseContext `json:"context,omitempty"`
	Event   AlexaEvent            `json:"event"`
}

type AlexaAuthorizationAcceptGrantInput struct {
	Grant struct {
		Type string `json:"type"`
		Code string `json:"code"`
	} `json:"grant"`
	Grantee TypeAndToken `json:"grantee"`
}

type AlexaDiscoveryDiscoverInput struct {
	Scope TypeAndToken `json:"scope"`
}

type AlexaReportStateInput NoPayload
type AlexaPowerControllerTurnOn NoPayload
type AlexaPowerControllerTurnOff NoPayload

type AlexaPlaybackControllerPlay NoPayload
type AlexaPlaybackControllerPause NoPayload
type AlexaPlaybackControllerStop NoPayload
type AlexaPlaybackControllerStartOver NoPayload
type AlexaPlaybackControllerPrevious NoPayload
type AlexaPlaybackControllerNext NoPayload
type AlexaPlaybackControllerRewind NoPayload
type AlexaPlaybackControllerFastForward NoPayload

type AlexaBrightnessControllerSetBrightness struct {
	Brightness int `json:"brightness"`
}

type AlexaPercentageControllerSetPercentage struct {
	Percentage int `json:"percentage"`
}

type AlexaColorTemperatureControllerSetColorTemperature struct {
	ColorTemperatureInKelvin int `json:"colorTemperatureInKelvin"`
}

type AlexaColorControllerSetColor struct {
	Color ColorHsv `json:"color"`
}

// in HSV
type ColorHsv struct {
	Hue        float64 `json:"hue"`
	Saturation float64 `json:"saturation"`
	Brightness float64 `json:"brightness"`
}

type EndpointsPayload struct {
	Endpoints []EndpointSpec `json:"endpoints"`
}

type ChangeReport struct {
	Change Change `json:"change"`
}

type Change struct {
	Properties []AlexaProperty `json:"properties"`
	Cause      *ChangeCause    `json:"cause,omitempty"`
}

type ChangeCause struct {
	Type string `json:"type"`
}

type EndpointSpec struct {
	EndpointId        string                   `json:"endpointId"`
	ManufacturerName  string                   `json:"manufacturerName"`
	Version           string                   `json:"version"`
	FriendlyName      string                   `json:"friendlyName"`
	Description       string                   `json:"description"`
	DisplayCategories []string                 `json:"displayCategories"`
	Capabilities      []EndpointSpecCapability `json:"capabilities"`
	Cookie            map[string]string        `json:"cookie"`
}

type EndpointSpecCapability struct {
	Type                string                            `json:"type"`
	Interface           string                            `json:"interface"`
	Version             string                            `json:"version"`
	Properties          *EndpointSpecCapabilityProperties `json:"properties,omitempty"`
	SupportedOperations []string                          `json:"supportedOperations,omitempty"`
}

type EndpointSpecCapabilityProperties struct {
	Supported           []EndpointSpecCapabilityNamedProperty `json:"supported"`
	ProactivelyReported bool                                  `json:"proactivelyReported"`
	Retrievable         bool                                  `json:"retrievable"`
}

type EndpointSpecCapabilityNamedProperty struct {
	Name string `json:"name"`
}
