package hapitypes

type LinkQualityEvent struct {
	Device      string
	LinkQuality uint // 0-100 %
}

func NewLinkQualityEvent(deviceId string, linkQuality uint) *LinkQualityEvent {
	return &LinkQualityEvent{
		Device:      deviceId,
		LinkQuality: linkQuality,
	}
}

func (e *LinkQualityEvent) InboundEventType() string {
	return "LinkQualityEvent"
}
