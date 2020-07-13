package ikeatradfri

import (
	"errors"
	"fmt"

	"github.com/lucasb-eyer/go-colorful"
)

// TODO: use native Golang COAP + DTLS to get rid of "$ coap-client" dependency which you manually have to compile

// 5712 means transition time
const (
	turnOnMsg  = `{ "3311": [{ "5850": 1 }] }`
	turnOffMsg = `{ "3311": [{ "5850": 0 }] }`
	dim        = `{ "3311": [{ "5851": %d, "5712": 5 }] }`
	colorMsg   = `{ "3311": [{ "5709": %d, "5710": %d, "5712": 5 }] }`
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

func SetRGB(deviceId string, r uint8, g uint8, b uint8, client *CoapClient) error {
	x, y := rgbToXY(r, g, b)

	return client.Put(
		deviceEndpoint(deviceId),
		fmt.Sprintf(colorMsg, x, y))
}

func SetColorTemp(deviceId string, kelvin uint, client *CoapClient) error {
	// TODO: find out if Trådfri actually supports only these warm/normal/cold?
	temp := tempConstantFromKelvin(kelvin)
	msg := fmt.Sprintf(colorMsg, colorTempX(temp), colorTempY(temp))

	return client.Put(deviceEndpoint(deviceId), msg)
}

func Dim(deviceId string, to int, client *CoapClient) error {
	if to < 0 || to > 254 {
		return errors.New("invalid argument")
	}

	return client.Put(
		deviceEndpoint(deviceId),
		fmt.Sprintf(dim, to))
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

// adapted from:
//   https://github.com/ffleurey/ThingML-Tradfri/blob/e4c8a8bbe48e36d364b9fd1c574e6a1e292bb58e/tradfri-java/src/main/java/org/thingml/tradfri/LightBulb.java#L106
func rgbToXY(red uint8, green uint8, blue uint8) (int, int) {
	// convert to XYZ
	x, y, z := colorful.Color{
		R: float64(red) / 255,
		G: float64(green) / 255,
		B: float64(blue) / 255,
	}.Xyz()

	// merge XYZ to XY (this seems really retarded)
	xMerged := x / (x + y + z)
	yMerged := y / (x + y + z)

	xInt := int(xMerged*65535 + 0.5)
	yInt := int(yMerged*65535 + 0.5)

	return xInt, yInt
}

func tempConstantFromKelvin(kelvin uint) ColorTemp {
	// https://developer.amazon.com/docs/device-apis/alexa-colortemperaturecontroller.html#setcolortemperature
	if kelvin < 4000 {
		return ColorTempWarm
	}

	if kelvin < 7000 {
		return ColorTempNormal
	}

	return ColorTempCold
}
