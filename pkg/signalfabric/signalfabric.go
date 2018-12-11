package signalfabric

import (
	"github.com/function61/home-automation-hub/pkg/hapitypes"
)

type Fabric struct {
	Event chan hapitypes.InboundEvent
}

func New() *Fabric {
	return &Fabric{
		Event: make(chan hapitypes.InboundEvent, 32),
	}
}

func (f *Fabric) Receive(e hapitypes.InboundEvent) {
	// TODO: log if channel full?
	f.Event <- e
}
