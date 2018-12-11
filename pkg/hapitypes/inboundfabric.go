package hapitypes

type InboundFabric struct {
	Ch chan InboundEvent
}

func NewInboundFabric() *InboundFabric {
	return &InboundFabric{
		Ch: make(chan InboundEvent, 32),
	}
}

func (f *InboundFabric) Receive(e InboundEvent) {
	// TODO: log if channel full?
	f.Ch <- e
}
