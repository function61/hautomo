package hapitypes

type CoverPositionEvent struct {
	DeviceId string
	Position uint // 0-100 %
}

var _ InboundEvent = (*CoverPositionEvent)(nil)
var _ OutboundEvent = (*CoverPositionEvent)(nil)

func NewCoverPositionEvent(deviceId string, position uint) *CoverPositionEvent {
	return &CoverPositionEvent{deviceId, position}
}

func (e *CoverPositionEvent) InboundEventType() string {
	return "CoverPositionEvent"
}

func (e *CoverPositionEvent) OutboundEventType() string {
	return "CoverPositionEvent"
}

func (e *CoverPositionEvent) RedirectInbound(toDeviceId string) InboundEvent {
	return NewCoverPositionEvent(toDeviceId, e.Position)
}
