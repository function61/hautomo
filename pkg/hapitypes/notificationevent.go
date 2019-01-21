package hapitypes

type NotificationEvent struct {
	Device  string
	Message string
}

func NewNotificationEvent(device string, message string) *NotificationEvent {
	return &NotificationEvent{device, message}
}

func (e *NotificationEvent) InboundEventType() string {
	return "NotificationEvent"
}

func (e *NotificationEvent) OutboundEventType() string {
	return "NotificationEvent"
}

func (e *NotificationEvent) RedirectInbound(toDeviceId string) InboundEvent {
	return NewNotificationEvent(toDeviceId, e.Message)
}
