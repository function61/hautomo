package signalfabric

import (
	"github.com/function61/home-automation-hub/pkg/hapitypes"
)

type Fabric struct {
	InfraredEvent   chan hapitypes.InfraredEvent
	PowerEvent      chan hapitypes.PowerEvent
	ColorEvent      chan hapitypes.ColorMsg
	BrightnessEvent chan hapitypes.BrightnessEvent
	PlaybackEvent   chan hapitypes.PlaybackEvent
}

func New() *Fabric {
	return &Fabric{
		InfraredEvent:   make(chan hapitypes.InfraredEvent, 1),
		PowerEvent:      make(chan hapitypes.PowerEvent, 1),
		ColorEvent:      make(chan hapitypes.ColorMsg, 1),
		BrightnessEvent: make(chan hapitypes.BrightnessEvent, 1),
		PlaybackEvent:   make(chan hapitypes.PlaybackEvent, 1),
	}
}
