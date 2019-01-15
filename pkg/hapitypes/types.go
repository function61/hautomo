package hapitypes

import (
	"errors"
	"github.com/function61/gokit/logex"
	"log"
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

func NewRGB(red, green, blue uint8) RGB {
	return RGB{
		Red:   red,
		Green: green,
		Blue:  blue,
	}
}

func (r RGB) IsGrayscale() bool {
	return r.Red == r.Green && r.Green == r.Blue
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
	Conf     AdapterConfig
	inbound  *InboundFabric     // inbound events coming from sensors, infrared, Amazon Echo etc.
	Outbound chan OutboundEvent // outbound events going to lights, TV, amplifier etc.
	Logl     *logex.Leveled
	confFile *ConfigFile // FIXME
}

func NewAdapter(conf AdapterConfig, confFile *ConfigFile, inbound *InboundFabric, logger *log.Logger) *Adapter {
	return &Adapter{
		Conf:     conf,
		inbound:  inbound,
		Outbound: make(chan OutboundEvent, 32),
		Logl:     logex.Levels(logger),
		confFile: confFile,
	}
}

// FIXME: remove the need for this
func (a *Adapter) GetConfigFileDeprecated() *ConfigFile {
	return a.confFile
}

func (a *Adapter) Send(e OutboundEvent) {
	// TODO: log warning if queue full?
	a.Outbound <- e
}

func (a *Adapter) Receive(e InboundEvent) {
	a.inbound.Receive(e)
}

func (a *Adapter) LogUnsupportedEvent(e OutboundEvent) {
	a.Logl.Error.Printf("unsupported outbound event: " + e.OutboundEventType())
}
