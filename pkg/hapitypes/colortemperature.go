package hapitypes

func NewColorTemperatureEvent(device string, temperatureInKelvin uint) *ColorTemperatureEvent {
	return &ColorTemperatureEvent{device, temperatureInKelvin}
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

func (e *ColorTemperatureEvent) RedirectInbound(toDeviceId string) InboundEvent {
	return NewColorTemperatureEvent(toDeviceId, e.TemperatureInKelvin)
}
