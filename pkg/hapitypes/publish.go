package hapitypes

type PublishEvent struct {
	Event string
}

func NewPublishEvent(deviceId string) *PublishEvent {
	return &PublishEvent{deviceId}
}

func (e *PublishEvent) InboundEventType() string {
	return "PublishEvent"
}
