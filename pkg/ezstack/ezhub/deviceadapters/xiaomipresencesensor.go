package deviceadapters

func init() {
	defineAdapter(modelAqaraPresenceSensor,
		aqaraVoltageEtc,
		withBatteryType(BatteryCR2032),
	)
}
