package hapitypes

import (
	"errors"
	"github.com/function61/gokit/logger"
)

/*
	symmetric events (same struct for inbound/outbound):

	ColorTemperatureEvent
	ColorMsg
	PersonPresenceChangeEvent
	PlaybackEvent

	asymmetric (different structs for inbound/outbound):

	inbound 							outbound
	--------------------------------------------
	PowerEvent							PowerMsg
	InfraredEvent						InfraredMsg
	BrightnessEvent						BrightnessMsg
*/

type OutboundEvent interface {
	OutboundEventType() string
}

type InboundEvent interface {
	InboundEventType() string
}

type RGB struct {
	Red   uint8
	Green uint8
	Blue  uint8
}

var ErrDeviceNotFound = errors.New("device not found")

type Device struct {
	Conf DeviceConfig

	// probably turned on if true
	// might be turned on even if false,
	ProbablyTurnedOn bool

	LastColor RGB
}

func NewDevice(conf DeviceConfig) *Device {
	return &Device{
		Conf: conf,

		// state
		ProbablyTurnedOn: false,
		LastColor:        RGB{255, 255, 255},
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

type Adapter struct {
	Id    string
	Event chan OutboundEvent
}

func NewAdapter(id string) *Adapter {
	return &Adapter{
		Id:    id,
		Event: make(chan OutboundEvent, 32),
	}
}

func (a *Adapter) Send(e OutboundEvent) {
	// TODO: log warning if queue full?
	a.Event <- e
}

func (a *Adapter) LogUnsupportedEvent(e OutboundEvent, log *logger.Logger) {
	log.Error("unsupported outbound event: " + e.OutboundEventType())
}
