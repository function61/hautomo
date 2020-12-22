package hapitypes

type SpeakEvent struct {
	Device  string
	Message string
}

func NewSpeakEvent(deviceId string, message string) *SpeakEvent {
	return &SpeakEvent{deviceId, message}
}

func (e *SpeakEvent) InboundEventType() string {
	return "SpeakEvent"
}

type PlaySoundEvent struct {
	Device string
	Url    string // sound file to play (at least support mp3). must be accessible via raw HTTP GET
}

func NewPlaySoundEvent(deviceId string, message string) *PlaySoundEvent {
	return &PlaySoundEvent{deviceId, message}
}

func (e *PlaySoundEvent) InboundEventType() string {
	return "PlaySoundEvent"
}

func (e *PlaySoundEvent) OutboundEventType() string {
	return "PlaySoundEvent"
}

func (e *PlaySoundEvent) RedirectInbound(toDeviceId string) InboundEvent {
	return NewPlaySoundEvent(toDeviceId, e.Url)
}
