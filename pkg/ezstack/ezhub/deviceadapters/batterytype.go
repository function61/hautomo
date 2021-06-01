package deviceadapters

import "github.com/function61/hautomo/pkg/ezstack/ezhub/hubtypes"

var (
	// "3V_2100" in zigbee2mqtt terminology
	BatteryCR2032 = &hubtypes.BatteryType{func(voltage float64) float64 {
		// curve taken from https://github.com/Koenkk/zigbee-herdsman-converters/blob/bc4314dea7f61a0c39daaa6d44c8d5afb8202ad4/lib/utils.js#L132
		return float64(func() float64 {
			voltage := float64(voltage * 1000) // [mV]

			switch {
			case voltage < 2100:
				return 0
			case voltage < 2440:
				return 6 - ((2440-voltage)*6)/340
			case voltage < 2740:
				return 18 - ((2740-voltage)*12)/300
			case voltage < 2900:
				return 42 - ((2900-voltage)*24)/160
			case voltage < 3000:
				return 100 - ((3000-voltage)*58)/100
			default:
				return 100
			}
		}()) / 100

	}}
)
