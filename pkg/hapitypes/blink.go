package hapitypes

type BlinkEvent struct {
	DeviceId string
}

func NewBlinkEvent(deviceId string) *BlinkEvent {
	return &BlinkEvent{deviceId}
}

func (e *BlinkEvent) InboundEventType() string {
	return "BlinkEvent"
}

func (e *BlinkEvent) OutboundEventType() string {
	return "BlinkEvent"
}
