package hubtypes

type BatteryType struct {
	// it'd be nice if we could just define simple linear (empty, full) voltage ranges,
	// but in reality it's a curve and thus different battery types need functions map points on curve
	// to the percentage of energy left. example curve for CR2032:
	// https://components101.com/asset/sites/default/files/inline-images/CR2032-Discharge-Time.png
	ToVoltage func(voltage float64) float64
}

// returns clamped range [0.0, 1.0]
func (b BatteryType) VoltageToPercentage(voltage float64) float64 {
	return b.ToVoltage(voltage)
}
