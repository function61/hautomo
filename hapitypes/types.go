package hapitypes

import (
	"errors"
)

type RGB struct {
	Red   uint8
	Green uint8
	Blue  uint8
}

type InfraredEvent struct {
	Remote string
	Event  string
}

func NewInfraredEvent(remote string, event string) InfraredEvent {
	return InfraredEvent{
		Remote: remote,
		Event:  event,
	}
}

type BrightnessEvent struct {
	DeviceIdOrDeviceGroupId string
	Brightness              uint
}

func NewBrightnessEvent(deviceIdOrDeviceGroupId string, brightness uint) BrightnessEvent {
	return BrightnessEvent{
		DeviceIdOrDeviceGroupId: deviceIdOrDeviceGroupId,
		Brightness:              brightness,
	}
}

type PowerKind int

const (
	PowerKindOn     PowerKind = iota
	PowerKindOff              = iota
	PowerKindToggle           = iota
)

type PowerEvent struct {
	DeviceIdOrDeviceGroupId string
	Kind                    PowerKind
}

func NewPowerEvent(deviceIdOrDeviceGroupId string, kind PowerKind) PowerEvent {
	return PowerEvent{
		DeviceIdOrDeviceGroupId: deviceIdOrDeviceGroupId,
		Kind: kind,
	}
}

func NewPowerToggleEvent(deviceIdOrDeviceGroupId string) PowerEvent {
	return PowerEvent{
		DeviceIdOrDeviceGroupId: deviceIdOrDeviceGroupId,
		Kind: PowerKindToggle,
	}
}

var ErrDeviceNotFound = errors.New("device not found")

type Device struct {
	Id          string
	Name        string
	Description string

	// adapter details
	AdapterId        string
	AdaptersDeviceId string // id by which the adapter identifies this device

	// probably turned on if true
	// might be turned on even if false,
	ProbablyTurnedOn bool

	PowerOnCmd  string
	PowerOffCmd string

	LastColor RGB
}

func NewDevice(id string, adapterId string, adaptersDeviceId string, name string, description string, powerOnCmd string, powerOffCmd string) *Device {
	return &Device{
		Id:          id,
		Name:        name,
		Description: description,

		AdapterId:        adapterId,
		AdaptersDeviceId: adaptersDeviceId,

		// state
		ProbablyTurnedOn: false,
		LastColor:        RGB{255, 255, 255},

		PowerOnCmd:  powerOnCmd,
		PowerOffCmd: powerOffCmd,
	}
}

type DeviceGroup struct {
	Id        string
	Name      string
	DeviceIds []string
}

func NewDeviceGroup(id string, name string, deviceIds []string) *DeviceGroup {
	return &DeviceGroup{
		Id:        id,
		Name:      name,
		DeviceIds: deviceIds,
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

type ColorMsg struct {
	DeviceId string
	Color    RGB
}

func NewColorMsg(deviceId string, color RGB) ColorMsg {
	return ColorMsg{
		DeviceId: deviceId,
		Color:    color,
	}
}

type InfraredMsg struct {
	DeviceId string // adapter's own id
	Command  string
}

func NewInfraredMsg(deviceId string, command string) InfraredMsg {
	return InfraredMsg{
		DeviceId: deviceId,
		Command:  command,
	}
}

type Adapter struct {
	Id          string
	PowerMsg    chan PowerMsg
	ColorMsg    chan ColorMsg
	InfraredMsg chan InfraredMsg
}

func NewAdapter(id string) *Adapter {
	return &Adapter{
		Id:          id,
		PowerMsg:    make(chan PowerMsg),
		ColorMsg:    make(chan ColorMsg),
		InfraredMsg: make(chan InfraredMsg),
	}
}
