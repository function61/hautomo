package hapitypes

type HeartbeatEvent struct {
	Device string
}

func NewHeartbeatEvent(deviceId string) *HeartbeatEvent {
	return &HeartbeatEvent{deviceId}
}

func (e *HeartbeatEvent) InboundEventType() string {
	return "HeartbeatEvent"
}
