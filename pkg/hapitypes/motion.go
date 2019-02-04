package hapitypes

type MotionEvent struct {
	Device      string
	Movement    bool
	Illuminance uint
}

func NewMotionEvent(deviceId string, movement bool, illuminance uint) *MotionEvent {
	return &MotionEvent{
		Device:      deviceId,
		Movement:    movement,
		Illuminance: illuminance,
	}
}

func (e *MotionEvent) InboundEventType() string {
	return "MotionEvent"
}
