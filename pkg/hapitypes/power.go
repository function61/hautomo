package hapitypes

type PowerKind int

const (
	PowerKindOn PowerKind = iota
	PowerKindOff
	PowerKindToggle
)

type PowerEvent struct {
	DeviceIdOrDeviceGroupId string
	Kind                    PowerKind
}

func (e *PowerEvent) InboundEventType() string {
	return "PowerEvent"
}

func NewPowerEvent(deviceIdOrDeviceGroupId string, kind PowerKind) PowerEvent {
	return PowerEvent{
		DeviceIdOrDeviceGroupId: deviceIdOrDeviceGroupId,
		Kind:                    kind,
	}
}

func NewPowerToggleEvent(deviceIdOrDeviceGroupId string) PowerEvent {
	return PowerEvent{
		DeviceIdOrDeviceGroupId: deviceIdOrDeviceGroupId,
		Kind:                    PowerKindToggle,
	}
}

type PowerMsg struct {
	DeviceId     string
	PowerCommand string
	On           bool
}

func NewPowerMsg(deviceId string, powerCommand string, on bool) PowerMsg {
	return PowerMsg{
		DeviceId:     deviceId,
		PowerCommand: powerCommand,
		On:           on,
	}
}

func (e *PowerMsg) OutboundEventType() string {
	return "PowerMsg"
}
