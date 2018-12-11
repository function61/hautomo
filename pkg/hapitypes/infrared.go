package hapitypes

type InfraredEvent struct {
	Remote string
	Event  string
}

func NewInfraredEvent(remote string, event string) InfraredEvent {
	return InfraredEvent{
		Remote: remote,
		Event:  event,
	}
}

func (e *InfraredEvent) InboundEventType() string {
	return "InfraredEvent"
}

type InfraredMsg struct {
	DeviceId string // adapter's own id
	Command  string
}

func NewInfraredMsg(deviceId string, command string) InfraredMsg {
	return InfraredMsg{
		DeviceId: deviceId,
		Command:  command,
	}
}

func (e *InfraredMsg) OutboundEventType() string {
	return "InfraredMsg"
}
