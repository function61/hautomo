package alexaadapter

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/function61/gokit/cryptorandombytes"
	"github.com/function61/gokit/ezhttp"
	"github.com/function61/hautomo/pkg/alexatypes"
)

func sendContactSensorEvent(
	ctx context.Context,
	endpointId string,
	contactClosed bool,
	alexaUserClient *http.Client,
) error {
	timeOfSample := time.Now()

	contactStr := func() string {
		if contactClosed {
			return "NOT_DETECTED"
		} else {
			return "DETECTED" // detected means "open" detected, not "contact" detected 🤦‍♀️
		}
	}()

	payloadJson, err := json.Marshal(alexatypes.ChangeReport{
		Change: alexatypes.Change{
			Properties: []alexatypes.AlexaProperty{
				{
					Namespace:                 "Alexa.ContactSensor",
					Name:                      "detectionState",
					Value:                     contactStr,
					TimeOfSample:              timeOfSample,
					UncertaintyInMilliseconds: 0,
				},
			},
		},
	})
	if err != nil {
		return err
	}
	payload := json.RawMessage(payloadJson)

	// TODO: add this as an actual type? (duplicated anyway)
	connectivityOk := struct {
		Value string `json:"value"`
	}{
		Value: "OK",
	}

	event := alexatypes.AlexaResponse{
		Event: alexatypes.AlexaEvent{
			Header: alexatypes.AlexaEventHeader{
				Namespace:      "Alexa",
				Name:           "ChangeReport",
				PayloadVersion: "3",
				MessageId:      cryptorandombytes.Hex(8),
			},
			Endpoint: &alexatypes.Endpoint{
				// docs say "scope" should be defined, but it works without it.. building
				// machinery to include it here isn't worth it
				EndpointId: endpointId,
			},
			Payload: &payload,
		},
		Context: &alexatypes.AlexaResponseContext{
			Properties: []alexatypes.AlexaProperty{
				{
					Namespace:                 "Alexa.EndpointHealth",
					Name:                      "connectivity",
					Value:                     connectivityOk,
					TimeOfSample:              timeOfSample,
					UncertaintyInMilliseconds: 0,
				},
			},
		},
	}

	_, err = ezhttp.Post(
		ctx,
		"https://api.amazonalexa.com/v3/events",
		ezhttp.Client(alexaUserClient),
		ezhttp.SendJson(event))
	return err
}
