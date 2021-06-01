package deviceadapters

func init() {
	defineAdapter(modelAqaraTemperatureSensor,
		aqaraVoltageEtc,
		withBatteryType(BatteryCR2032),
	)
}
