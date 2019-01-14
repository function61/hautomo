package hapitypes

type PushButtonEvent struct {
	Device    string
	Specifier string // single/double/...
}

func NewPushButtonEvent(deviceId string, specifier string) *PushButtonEvent {
	return &PushButtonEvent{
		Device:    deviceId,
		Specifier: specifier,
	}
}

func (e *PushButtonEvent) InboundEventType() string {
	return "PushButtonEvent"
}
