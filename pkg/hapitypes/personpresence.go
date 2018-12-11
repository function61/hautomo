package hapitypes

func NewPersonPresenceChangeEvent(personId string, present bool) PersonPresenceChangeEvent {
	return PersonPresenceChangeEvent{
		PersonId: personId,
		Present:  present,
	}
}

type PersonPresenceChangeEvent struct {
	PersonId string
	Present  bool
}

func (e *PersonPresenceChangeEvent) InboundEventType() string {
	return "PersonPresenceChangeEvent"
}
