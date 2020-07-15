// Messages with which the adapter and the connector communicate by
package aamessages

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

type TurnOnRequest struct {
	DeviceIdOrDeviceGroupId string `json:"id"`
}

func (x *TurnOnRequest) Kind() string { return "turn_on" }

type TurnOffRequest struct {
	DeviceIdOrDeviceGroupId string `json:"id"`
}

func (x *TurnOffRequest) Kind() string { return "turn_off" }

type ColorRequest struct {
	DeviceIdOrDeviceGroupId string `json:"id"`
	Red                     uint8  `json:"red"`
	Green                   uint8  `json:"green"`
	Blue                    uint8  `json:"blue"`
}

func (x *ColorRequest) Kind() string { return "color" }

type BrightnessRequest struct {
	DeviceIdOrDeviceGroupId string `json:"id"`
	Brightness              uint   `json:"brightness"` // 0-100
}

func (x *BrightnessRequest) Kind() string { return "brightness" }

type PlaybackRequest struct {
	DeviceIdOrDeviceGroupId string `json:"id"`
	Action                  string `json:"action"`
}

func (x *PlaybackRequest) Kind() string { return "playback" }

type ColorTemperatureRequest struct {
	DeviceIdOrDeviceGroupId  string `json:"id"`
	ColorTemperatureInKelvin uint   `json:"colorTemperatureInKelvin"`
}

func (x *ColorTemperatureRequest) Kind() string { return "colorTemperature" }

type Message interface {
	Kind() string
}

var allocators = map[string]func() Message{
	"turn_on":          func() Message { return &TurnOnRequest{} },
	"turn_off":         func() Message { return &TurnOffRequest{} },
	"color":            func() Message { return &ColorRequest{} },
	"brightness":       func() Message { return &BrightnessRequest{} },
	"playback":         func() Message { return &PlaybackRequest{} },
	"colorTemperature": func() Message { return &ColorTemperatureRequest{} },
}

func Marshal(msg Message) (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s %s", msg.Kind(), jsonBytes), nil
}

var parseMessageRegexp = regexp.MustCompile(`^([a-zA-Z_0-9]+) (.+)$`)

func Unmarshal(input string) (Message, error) {
	match := parseMessageRegexp.FindStringSubmatch(input)
	if match == nil {
		return nil, errors.New("Unmarshal(): invalid format")
	}

	kind := match[1]

	allocator, found := allocators[kind]
	if !found {
		return nil, fmt.Errorf("invalid kind for allocator: %s", kind)
	}

	msg := allocator()

	return msg, json.Unmarshal([]byte(match[2]), msg)
}
