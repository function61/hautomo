package zigbee2mqttadapter

import (
	"math"
)

// https://en.wikipedia.org/wiki/Mired
// https://developer.amazon.com/docs/device-apis/alexa-colortemperaturecontroller.html#setcolortemperature
// https://github.com/Koenkk/zigbee2mqtt/issues/627
func kelvinToMired(k uint) uint {
	return 1000000 / k
}

// stolen from https://github.com/lindsaymarkward/go-yeelight/blob/master/yeelight.go
func temperatureToRGB(kelvin float64) (r, g, b uint8) {
	temp := kelvin / 100

	var red, green, blue float64

	if temp <= 66 {

		red = 255

		green = temp
		green = 99.4708025861*math.Log(green) - 161.1195681661

		if temp <= 19 {
			blue = 0
		} else {
			blue = temp - 10
			blue = 138.5177312231*math.Log(blue) - 305.0447927307
		}

	} else {
		red = temp - 60
		red = 329.698727446 * math.Pow(red, -0.1332047592)

		green = temp - 60
		green = 288.1221695283 * math.Pow(green, -0.0755148492)

		blue = 255
	}
	return clamp(red, 0, 255), clamp(green, 0, 255), clamp(blue, 0, 255)
}

func clamp(x, min, max float64) uint8 {
	if x < min {
		return uint8(min)
	}
	if x > max {
		return uint8(max)
	}
	return uint8(x)
}
