package hapitypes

type WaterLeakEvent struct {
	Device        string
	WaterDetected bool
}

func NewWaterLeakEvent(deviceId string, waterDetected bool) *WaterLeakEvent {
	return &WaterLeakEvent{deviceId, waterDetected}
}

func (e *WaterLeakEvent) InboundEventType() string {
	return "WaterLeakEvent"
}
