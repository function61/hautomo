package zigbee2mqttadapter

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/function61/hautomo/pkg/hapitypes"
)

type resolvedDevice struct {
	id   string // not adapter's device id, but internal id
	kind deviceKind
}

type deviceResolver func(deviceId string) *resolvedDevice

func parseMsgPayload(topicName string, resolver deviceResolver, message string, now time.Time) ([]hapitypes.InboundEvent, error) {
	// block "zigbee2mqtt/0x00158d000227a73c/set", which is probably publishes made by us
	if strings.HasSuffix(topicName, "/set") {
		return nil, nil
	}

	// "zigbee2mqtt/0x00158d000227a73c" => "0x00158d000227a73c"
	foreignId := topicName[len(z2mTopicPrefix):]

	if foreignId == "bridge/log" { // bridge's log messages
		// prevent log filling with "device bridge/log unrecognized" messages
		return nil, nil
	}

	resolved := resolver(foreignId)
	if resolved == nil {
		return nil, fmt.Errorf("device %s unrecognized", foreignId)
	}

	ourId := resolved.id

	events := []hapitypes.InboundEvent{}
	push := func(e hapitypes.InboundEvent) {
		events = append(events, e)
	}

	switch resolved.kind {
	case deviceKindRTCGQ11LM:
		payload := RTCGQ11LM{}
		if err := decJson(&payload, message); err != nil {
			return nil, err
		}

		push(hapitypes.NewMotionEvent(ourId, payload.Occupancy, payload.Illuminance))
		push(hapitypes.NewLinkQualityEvent(ourId, payload.LinkQuality))
	case deviceKindDJT11LM:
		payload := DJT11LM{}
		if err := decJson(&payload, message); err != nil {
			return nil, err
		}

		if payload.Action != "" {
			push(hapitypes.NewVibrationEvent(ourId))
		}

		push(hapitypes.NewLinkQualityEvent(ourId, payload.LinkQuality))
		push(hapitypes.NewBatteryStatusEvent(ourId, payload.Battery, payload.Voltage))
	case deviceKindWXKG11LM:
		payload := WXKG11LM{}
		if err := decJson(&payload, message); err != nil {
			return nil, err
		}

		if payload.Action != nil {
			push(hapitypes.NewPushButtonEvent(ourId, *payload.Action))
		}

		push(hapitypes.NewLinkQualityEvent(ourId, payload.LinkQuality))
		push(hapitypes.NewBatteryStatusEvent(ourId, payload.Battery, payload.Voltage))
	case deviceKindWXKG02LM:
		payload := WXKG02LM{}
		if err := decJson(&payload, message); err != nil {
			return nil, err
		}

		if payload.Action != nil {
			push(hapitypes.NewPushButtonEvent(ourId, *payload.Action))
		}

		push(hapitypes.NewLinkQualityEvent(ourId, payload.LinkQuality))
	case deviceKindMCCGQ11LM:
		payload := MCCGQ11LM{}
		if err := decJson(&payload, message); err != nil {
			return nil, err
		}

		push(hapitypes.NewContactEvent(ourId, payload.Contact, now))

		push(hapitypes.NewLinkQualityEvent(ourId, payload.LinkQuality))
		push(hapitypes.NewBatteryStatusEvent(ourId, payload.Battery, payload.Voltage))
	case deviceKindSJCGQ11LM:
		payload := SJCGQ11LM{}
		if err := decJson(&payload, message); err != nil {
			return nil, err
		}

		push(hapitypes.NewWaterLeakEvent(ourId, payload.WaterLeak))

		push(hapitypes.NewLinkQualityEvent(ourId, payload.LinkQuality))
		push(hapitypes.NewBatteryStatusEvent(ourId, payload.Battery, payload.Voltage))
	case deviceKindWSDCGQ11LM:
		payload := WSDCGQ11LM{}
		if err := decJson(&payload, message); err != nil {
			return nil, err
		}

		push(hapitypes.NewTemperatureHumidityPressureEvent(
			ourId,
			payload.Temperature,
			payload.Humidity,
			payload.Pressure,
		))

		push(hapitypes.NewLinkQualityEvent(ourId, payload.LinkQuality))
		push(hapitypes.NewBatteryStatusEvent(ourId, payload.Battery, payload.Voltage))
	case deviceKindE1524:
		payload := E1524{}
		if err := decJson(&payload, message); err != nil {
			return nil, err
		}

		push(hapitypes.NewPushButtonEvent(ourId, payload.Action))
		push(hapitypes.NewLinkQualityEvent(ourId, payload.LinkQuality))
	case deviceKindUnknown:
		return nil, fmt.Errorf("unknown device kind for %s", ourId)
	default:
		return nil, fmt.Errorf("unsupported device kind for %s, %d", ourId, resolved.kind)
	}

	return events, nil
}

func decJson(ref interface{}, data string) error {
	return json.Unmarshal([]byte(data), ref)
}
