package hapitypes

type PublishEvent struct {
	Topic string
}

func NewPublishEvent(topic string) *PublishEvent {
	return &PublishEvent{topic}
}

func (e *PublishEvent) InboundEventType() string {
	return "PublishEvent"
}
