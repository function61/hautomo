package hapitypes

import (
	"time"
)

type Statefile struct {
	Devices map[string]DeviceStateSnapshot `json:"device_state_snapshots_by_id"`
}

func NewStatefile() Statefile {
	return Statefile{
		Devices: map[string]DeviceStateSnapshot{},
	}
}

// TODO: just compose device's state with this?
// TODO: LastTemperatureHumidityPressureEvent should have explicit JSON annotations
type DeviceStateSnapshot struct {
	ProbablyTurnedOn                     bool                              `json:"probably_turned_on"`
	LastColor                            RGB                               `json:"last_color"`
	LastTemperatureHumidityPressureEvent *TemperatureHumidityPressureEvent `json:"last_temperaturehumiditypressure"`
	LastOnline                           *time.Time                        `json:"last_online"`
	LinkQuality                          uint                              `json:"link_quality_pct"`
	BatteryPct                           uint                              `json:"battery_pct"`
	BatteryVoltage                       uint                              `json:"battery_voltage_mv"`
}

func (d *Device) SnapshotState() (*DeviceStateSnapshot, error) {
	return &DeviceStateSnapshot{
		LastColor:                            d.LastColor,
		LastTemperatureHumidityPressureEvent: d.LastTemperatureHumidityPressureEvent,
		LastOnline:                           d.LastOnline,
		LinkQuality:                          d.LinkQuality,
		BatteryPct:                           d.BatteryPct,
		BatteryVoltage:                       d.BatteryVoltage,
	}, nil
}

func (d *Device) RestoreStateFromSnapshot(snapshot DeviceStateSnapshot) error {
	d.LastColor = snapshot.LastColor
	d.LastTemperatureHumidityPressureEvent = snapshot.LastTemperatureHumidityPressureEvent
	d.LastOnline = snapshot.LastOnline
	d.LinkQuality = snapshot.LinkQuality
	d.BatteryPct = snapshot.BatteryPct
	d.BatteryVoltage = snapshot.BatteryVoltage

	return nil
}
