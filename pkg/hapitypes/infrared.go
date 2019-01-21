package hapitypes

type RawInfraredEvent struct {
	Remote string
	Event  string
}

func NewRawInfraredEvent(remote string, event string) *RawInfraredEvent {
	return &RawInfraredEvent{
		Remote: remote,
		Event:  event,
	}
}

func (e *RawInfraredEvent) InboundEventType() string {
	return "RawInfraredEvent"
}

type InfraredEvent struct {
	Device  string
	Command string
}

func NewInfraredEvent(device string, command string) *InfraredEvent {
	return &InfraredEvent{
		Device:  device,
		Command: command,
	}
}

func (e *InfraredEvent) InboundEventType() string {
	return "InfraredEvent"
}

func (e *InfraredEvent) OutboundEventType() string {
	return "InfraredEvent"
}

func (e *InfraredEvent) RedirectInbound(toDeviceId string) InboundEvent {
	return NewInfraredEvent(toDeviceId, e.Command)
}
