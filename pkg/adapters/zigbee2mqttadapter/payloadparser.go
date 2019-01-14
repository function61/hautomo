package zigbee2mqttadapter

import (
	"encoding/json"
	"fmt"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
)

type resolvedDevice struct {
	id   string // not adapter's device id, but internal id
	kind deviceKind
}

type deviceResolver func(deviceId string) *resolvedDevice

func parseMsgPayload(topicName string, resolver deviceResolver, message string) (hapitypes.InboundEvent, error) {
	// "zigbee2mqtt/0x00158d000227a73c" => "0x00158d000227a73c"
	foreignId := topicName[len(z2mTopicPrefix):]

	resolved := resolver(foreignId)
	if resolved == nil {
		return nil, fmt.Errorf("device %s unrecognized", foreignId)
	}

	ourId := resolved.id

	switch resolved.kind {
	case deviceKindWXKG11LM:
		payload := WXKG11LM{}
		if err := decJson(&payload, message); err != nil {
			return nil, err
		}

		if payload.Click == nil {
			return hapitypes.NewHeartbeatEvent(ourId), nil
		}

		return hapitypes.NewPushButtonEvent(ourId, *payload.Click), nil
	case deviceKindMCCGQ11LM:
		payload := MCCGQ11LM{}
		if err := decJson(&payload, message); err != nil {
			return nil, err
		}

		return hapitypes.NewContactEvent(ourId, payload.Contact), nil
	case deviceKindSJCGQ11LM:
		payload := SJCGQ11LM{}
		if err := decJson(&payload, message); err != nil {
			return nil, err
		}

		return hapitypes.NewWaterLeakEvent(ourId, payload.WaterLeak), nil
	case deviceKindWSDCGQ11LM:
		payload := WSDCGQ11LM{}
		if err := decJson(&payload, message); err != nil {
			return nil, err
		}

		return hapitypes.NewTemperatureHumidityPressureEvent(
			ourId,
			payload.Temperature,
			payload.Humidity,
			payload.Pressure,
		), nil
	case deviceKindUnknown:
		return nil, fmt.Errorf("unknown device kind for %s", ourId)
	default:
		return nil, fmt.Errorf("unsupported device kind for %s, %d", ourId, resolved.kind)
	}
}

func decJson(ref interface{}, data string) error {
	return json.Unmarshal([]byte(data), ref)
}
