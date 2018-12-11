package hapitypes

func NewColorTemperatureEvent(deviceIdOrDeviceGroupId string, temperatureInKelvin uint) ColorTemperatureEvent {
	return ColorTemperatureEvent{deviceIdOrDeviceGroupId, temperatureInKelvin}
}

type ColorTemperatureEvent struct {
	Device              string
	TemperatureInKelvin uint
}

func (e *ColorTemperatureEvent) InboundEventType() string {
	return "ColorTemperatureEvent"
}

func (e *ColorTemperatureEvent) OutboundEventType() string {
	return "ColorTemperatureEvent"
}
