package hapitypes

type PlaybackEvent struct {
	Device string
	Action string
}

func NewPlaybackEvent(device string, action string) *PlaybackEvent {
	return &PlaybackEvent{
		Device: device,
		Action: action,
	}
}

func (e *PlaybackEvent) InboundEventType() string {
	return "PlaybackEvent"
}

func (e *PlaybackEvent) OutboundEventType() string {
	return "PlaybackEvent"
}
