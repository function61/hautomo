package main

import (
	"errors"
)

type InfraredEvent struct {
	Remote string
	Event  string
}

type RGB struct {
	Red   uint8
	Green uint8
	Blue  uint8
}

func NewInfraredEvent(remote string, event string) InfraredEvent {
	return InfraredEvent{
		Remote: remote,
		Event:  event,
	}
}

type powerKind int

const (
	powerKindOn     powerKind = iota
	powerKindOff              = iota
	powerKindToggle           = iota
)

type PowerEvent struct {
	DeviceIdOrDeviceGroupId string
	Kind                    powerKind
}

func NewPowerEvent(deviceIdOrDeviceGroupId string, kind powerKind) PowerEvent {
	return PowerEvent{
		DeviceIdOrDeviceGroupId: deviceIdOrDeviceGroupId,
		Kind: kind,
	}
}

func NewPowerToggleEvent(deviceIdOrDeviceGroupId string) PowerEvent {
	return PowerEvent{
		DeviceIdOrDeviceGroupId: deviceIdOrDeviceGroupId,
		Kind: powerKindToggle,
	}
}

var errDeviceNotFound = errors.New("device not found")

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
