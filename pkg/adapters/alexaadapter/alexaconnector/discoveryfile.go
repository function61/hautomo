package alexaconnector

import (
	"encoding/json"
	"fmt"

	"github.com/function61/hautomo/pkg/adapters/alexaadapter/alexadevicesync"
	"github.com/function61/hautomo/pkg/alexatypes"
)

func discoveryFileToAlexaInterfaces(
	file alexadevicesync.AlexaConnectorSpec,
	msgIdGenerator messageIdGenerator,
) (*alexatypes.AlexaResponse, error) {
	endpoints := []alexatypes.EndpointSpec{}
	for _, dev := range append(file.Devices) {
		caps := []alexatypes.EndpointSpecCapability{}
		for _, capCode := range dev.CapabilityCodes {
			// some caps yield multiple Alexa capabilities, e.g. contact sensor
			// actually requires ("MUST") = ["Alexa.ContactSensor", "Alexa.EndpointHealth", "Alexa"]
			// while a power controller only requires = ["Alexa.PowerController"]
			capsToAdd, err := makeCaps(capCode)
			if err != nil {
				return nil, err
			}

			caps = append(caps, capsToAdd...)
		}

		endpoints = append(endpoints, alexatypes.EndpointSpec{
			EndpointId:        dev.Id,
			ManufacturerName:  "function61.com",
			Version:           "1.0",
			FriendlyName:      dev.FriendlyName,
			Description:       dev.Description,
			DisplayCategories: []string{dev.DisplayCategory},
			Capabilities:      caps,
			Cookie: map[string]string{
				"queue": file.Queue,
			},
		})
	}

	payloadJson, err := json.MarshalIndent(alexatypes.EndpointsPayload{Endpoints: endpoints}, "", "  ")
	if err != nil {
		return nil, err
	}
	payloadJsonRaw := json.RawMessage(payloadJson)

	return &alexatypes.AlexaResponse{
		Event: alexatypes.AlexaEvent{
			Header: alexatypes.AlexaEventHeader{
				Namespace:      "Alexa.Discovery",
				Name:           "Discover.Response",
				PayloadVersion: "3",
				MessageId:      msgIdGenerator(),
			},
			Payload: &payloadJsonRaw,
		},
	}, nil
}

func makeCaps(capCode string) ([]alexatypes.EndpointSpecCapability, error) {
	// helper for making slice
	one := func(x alexatypes.EndpointSpecCapability) []alexatypes.EndpointSpecCapability {
		return []alexatypes.EndpointSpecCapability{x}
	}

	switch capCode {
	case "PowerController":
		return one(capNonRetrievableNonProactivelyReported("Alexa.PowerController", "powerState")), nil
	case "BrightnessController":
		return one(capNonRetrievableNonProactivelyReported("Alexa.BrightnessController", "brightness")), nil
	case "ColorController":
		return one(capNonRetrievableNonProactivelyReported("Alexa.ColorController", "color")), nil
	case "ColorTemperatureController":
		return one(capNonRetrievableNonProactivelyReported("Alexa.ColorTemperatureController", "colorTemperatureInKelvin")), nil
	case "PercentageController":
		return one(capNonRetrievableNonProactivelyReported("Alexa.PercentageController", "percentage")), nil
	case "PlaybackController":
		return one(alexatypes.EndpointSpecCapability{
			Type:                "AlexaInterface",
			Interface:           "Alexa.PlaybackController",
			Version:             "3",
			SupportedOperations: []string{"Play", "Pause", "Stop"},
		}), nil
	default:
		return nil, fmt.Errorf("unsupported capability code: %s", capCode)
	}
}

func capNonRetrievableNonProactivelyReported(
	interfaceName string,
	supportedProperties ...string,
) alexatypes.EndpointSpecCapability {
	return capInternal(interfaceName, false, supportedProperties...)
}

func capRetrievableAndProactivelyReported(
	interfaceName string,
	supportedProperties ...string,
) alexatypes.EndpointSpecCapability {
	return capInternal(interfaceName, true, supportedProperties...)
}

func capInternal(
	interfaceName string,
	retrievableAndProactivelyReported bool,
	supportedProperties ...string,
) alexatypes.EndpointSpecCapability {
	supported := []alexatypes.EndpointSpecCapabilityNamedProperty{}
	for _, supportedProperty := range supportedProperties {
		supported = append(supported, alexatypes.EndpointSpecCapabilityNamedProperty{
			Name: supportedProperty,
		})
	}

	properties := func() *alexatypes.EndpointSpecCapabilityProperties {
		if len(supported) == 0 {
			return nil
		}

		return &alexatypes.EndpointSpecCapabilityProperties{
			Supported:           supported,
			ProactivelyReported: retrievableAndProactivelyReported,
			Retrievable:         retrievableAndProactivelyReported,
		}
	}()

	return alexatypes.EndpointSpecCapability{
		Type:       "AlexaInterface",
		Interface:  interfaceName,
		Version:    "3",
		Properties: properties,
	}
}

// I don't know what this actually does, nor why it's suggested in the docs
func plainAlexaInterface() alexatypes.EndpointSpecCapability {
	return alexatypes.EndpointSpecCapability{
		Type:      "AlexaInterface",
		Interface: "Alexa",
		Version:   "3",
	}
}
