package hapitypes

type VibrationEvent struct {
	Device string
}

func NewVibrationEvent(deviceId string) *VibrationEvent {
	return &VibrationEvent{
		Device: deviceId,
	}
}

func (e *VibrationEvent) InboundEventType() string {
	return "VibrationEvent"
}
