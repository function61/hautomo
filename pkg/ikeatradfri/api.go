package ikeatradfri

import (
	"errors"
	"fmt"
)

// TODO: use native Golang COAP + DTLS to get rid of "$ coap-client" dependency which you manually have to compile

const (
	turnOnMsg           = `{ "3311": [{ "5850": 1 }] }`
	turnOffMsg          = `{ "3311": [{ "5850": 0 }] }`
	dimWithoutFadingMsg = `{ "3311": [{ "5851": %d }] }`
	colorTemperature    = `{ "3311": [{ "5709": %d, "5710": %d }] }`
)

type ColorTemp int

const (
	ColorTempWarm ColorTemp = iota
	ColorTempNormal
	ColorTempCold
)

func TurnOn(deviceId string, client *CoapClient) error {
	return client.Put(deviceEndpoint(deviceId), turnOnMsg)
}

func TurnOff(deviceId string, client *CoapClient) error {
	return client.Put(deviceEndpoint(deviceId), turnOffMsg)
}

func SetColorTemp(deviceId string, temp ColorTemp, client *CoapClient) error {
	msg := fmt.Sprintf(colorTemperature, colorTempX(temp), colorTempY(temp))

	return client.Put(deviceEndpoint(deviceId), msg)
}

func DimWithoutFading(deviceId string, to int, client *CoapClient) error {
	if to < 0 || to > 254 {
		return errors.New("invalid argument")
	}

	return client.Put(
		deviceEndpoint(deviceId),
		fmt.Sprintf(dimWithoutFadingMsg, to))
}

func deviceEndpoint(deviceId string) string {
	return "/15001/" + deviceId
}

func colorTempX(temp ColorTemp) int {
	switch temp {
	case ColorTempCold:
		return 24930
	case ColorTempNormal:
		return 30140
	case ColorTempWarm:
		return 33135
	default:
		panic("unknown temp")
	}
}

func colorTempY(temp ColorTemp) int {
	switch temp {
	case ColorTempCold:
		return 24684
	case ColorTempNormal:
		return 26909
	case ColorTempWarm:
		return 27211
	default:
		panic("unknown temp")
	}
}
