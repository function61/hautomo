package hapitypes

import (
	"time"
)

type ContactEvent struct {
	Device  string
	Contact bool
	When    time.Time
}

func NewContactEvent(deviceId string, contact bool, now time.Time) *ContactEvent {
	return &ContactEvent{deviceId, contact, now}
}

func (e *ContactEvent) InboundEventType() string {
	return "ContactEvent"
}
