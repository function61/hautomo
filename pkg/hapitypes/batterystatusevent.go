package hapitypes

type BatteryStatusEvent struct {
	Device     string
	BatteryPct uint // 0-100 %
	Voltage    uint // [mV]
}

func NewBatteryStatusEvent(deviceId string, batteryPct uint, voltage uint) *BatteryStatusEvent {
	return &BatteryStatusEvent{
		Device:     deviceId,
		BatteryPct: batteryPct,
		Voltage:    voltage,
	}
}

func (e *BatteryStatusEvent) InboundEventType() string {
	return "BatteryStatusEvent"
}
