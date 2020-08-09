package hapitypes

import (
	"github.com/function61/gokit/logex"
)

type InboundFabric struct {
	Ch   chan InboundEvent
	logl *logex.Leveled
}

func NewInboundFabric(logl *logex.Leveled) *InboundFabric {
	return &InboundFabric{
		Ch:   make(chan InboundEvent, 32),
		logl: logl,
	}
}

func (f *InboundFabric) Receive(e InboundEvent) {
	select {
	case f.Ch <- e:
	default:
		f.logl.Error.Printf(
			"Inbound.Receive blocks because buffer (%d) is full. Unless main loop drains soon, expect severe problems.",
			cap(f.Ch))

		// don't drop messages, this will probably block for a while (or indefinitely, if
		// main loop is stuck)
		f.Ch <- e
	}
}
