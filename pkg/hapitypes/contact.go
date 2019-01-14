package hapitypes

type ContactEvent struct {
	Device  string
	Contact bool
}

func NewContactEvent(deviceId string, contact bool) *ContactEvent {
	return &ContactEvent{deviceId, contact}
}

func (e *ContactEvent) InboundEventType() string {
	return "ContactEvent"
}
