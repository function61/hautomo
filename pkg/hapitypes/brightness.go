package hapitypes

type BrightnessEvent struct {
	DeviceIdOrDeviceGroupId string
	Brightness              uint // 0..100 %
}

func NewBrightnessEvent(deviceIdOrDeviceGroupId string, brightness uint) *BrightnessEvent {
	return &BrightnessEvent{
		DeviceIdOrDeviceGroupId: deviceIdOrDeviceGroupId,
		Brightness:              brightness,
	}
}

func (e *BrightnessEvent) InboundEventType() string {
	return "BrightnessEvent"
}

type BrightnessMsg struct {
	DeviceId   string
	Brightness uint
	LastColor  RGB
}

func NewBrightnessMsg(deviceId string, brightness uint, lastColor RGB) *BrightnessMsg {
	return &BrightnessMsg{
		DeviceId:   deviceId,
		Brightness: brightness,
		LastColor:  lastColor,
	}
}

func (e *BrightnessMsg) OutboundEventType() string {
	return "BrightnessMsg"
}
