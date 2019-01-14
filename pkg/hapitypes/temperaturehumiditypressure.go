package hapitypes

type TemperatureHumidityPressureEvent struct {
	Device      string
	Temperature float64
	Humidity    float64
	Pressure    float64
}

func NewTemperatureHumidityPressureEvent(deviceId string, temperature float64, humidity float64, pressure float64) *TemperatureHumidityPressureEvent {
	return &TemperatureHumidityPressureEvent{
		Device:      deviceId,
		Temperature: temperature,
		Humidity:    humidity,
		Pressure:    pressure,
	}
}

func (e *TemperatureHumidityPressureEvent) InboundEventType() string {
	return "TemperatureHumidityPressureEvent"
}
