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
	// whether this was explicitly asked by the user, or generated (f.ex. by a device
	// group on => multiple ons for different devices)
	Explicit bool
}

func (e *PowerEvent) InboundEventType() string {
	return "PowerEvent"
}

func NewPowerEvent(deviceIdOrDeviceGroupId string, kind PowerKind, explicit bool) *PowerEvent {
	return &PowerEvent{
		DeviceIdOrDeviceGroupId: deviceIdOrDeviceGroupId,
		Kind:                    kind,
		Explicit:                explicit,
	}
}

func NewPowerToggleEvent(deviceIdOrDeviceGroupId string, explicit bool) PowerEvent {
	return PowerEvent{
		DeviceIdOrDeviceGroupId: deviceIdOrDeviceGroupId,
		Kind:                    PowerKindToggle,
		Explicit:                explicit,
	}
}

type PowerMsg struct {
	DeviceId     string
	PowerCommand string
	On           bool
}

func NewPowerMsg(deviceId string, powerCommand string, on bool) *PowerMsg {
	return &PowerMsg{
		DeviceId:     deviceId,
		PowerCommand: powerCommand,
		On:           on,
	}
}

func (e *PowerMsg) OutboundEventType() string {
	return "PowerMsg"
}

// used by device group adapter, therefore we can mark these as implicit
func (e *PowerMsg) RedirectInbound(toDeviceId string) InboundEvent {
	if e.On {
		return NewPowerEvent(toDeviceId, PowerKindOn, false)
	}
	return NewPowerEvent(toDeviceId, PowerKindOff, false)
}
