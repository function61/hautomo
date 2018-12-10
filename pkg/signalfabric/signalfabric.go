package signalfabric

import (
	"github.com/function61/home-automation-hub/pkg/hapitypes"
)

type Fabric struct {
	InfraredEvent             chan hapitypes.InfraredEvent
	PowerEvent                chan hapitypes.PowerEvent
	ColorEvent                chan hapitypes.ColorMsg
	BrightnessEvent           chan hapitypes.BrightnessEvent
	PlaybackEvent             chan hapitypes.PlaybackEvent
	ColorTemperatureEvent     chan hapitypes.ColorTemperatureEvent
	PersonPresenceChangeEvent chan hapitypes.PersonPresenceChangeEvent
}

func New() *Fabric {
	return &Fabric{
		InfraredEvent:             make(chan hapitypes.InfraredEvent, 1),
		PowerEvent:                make(chan hapitypes.PowerEvent, 1),
		ColorEvent:                make(chan hapitypes.ColorMsg, 1),
		BrightnessEvent:           make(chan hapitypes.BrightnessEvent, 1),
		PlaybackEvent:             make(chan hapitypes.PlaybackEvent, 1),
		ColorTemperatureEvent:     make(chan hapitypes.ColorTemperatureEvent, 1),
		PersonPresenceChangeEvent: make(chan hapitypes.PersonPresenceChangeEvent, 1),
	}
}
