package hapitypes

import (
	"time"
)

type ContactEvent struct {
	Device  string
	Contact bool
	When    time.Time
}

func NewContactEvent(deviceId string, contact bool) *ContactEvent {
	return &ContactEvent{deviceId, contact, time.Now()}
}

func (e *ContactEvent) InboundEventType() string {
	return "ContactEvent"
}
