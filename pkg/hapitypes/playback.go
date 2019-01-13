package hapitypes

type PlaybackEvent struct {
	DeviceIdOrDeviceGroupId string
	Action                  string
}

func NewPlaybackEvent(deviceIdOrDeviceGroupId string, action string) *PlaybackEvent {
	return &PlaybackEvent{
		DeviceIdOrDeviceGroupId: deviceIdOrDeviceGroupId,
		Action:                  action,
	}
}

func (e *PlaybackEvent) InboundEventType() string {
	return "PlaybackEvent"
}

func (e *PlaybackEvent) OutboundEventType() string {
	return "PlaybackEvent"
}
