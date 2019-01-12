package zigbee2mqttadapter

// https://en.wikipedia.org/wiki/Mired
// https://developer.amazon.com/docs/device-apis/alexa-colortemperaturecontroller.html#setcolortemperature
// https://github.com/Koenkk/zigbee2mqtt/issues/627
func kelvinToMired(k uint) uint {
	return 1000000 / k
}
