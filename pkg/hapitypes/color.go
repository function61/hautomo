package hapitypes

type ColorMsg struct {
	DeviceId string
	Color    RGB
}

func NewColorMsg(deviceId string, color RGB) *ColorMsg {
	return &ColorMsg{
		DeviceId: deviceId,
		Color:    color,
	}
}

func (e *ColorMsg) InboundEventType() string {
	return "ColorMsg"
}

func (e *ColorMsg) OutboundEventType() string {
	return "ColorMsg"
}
