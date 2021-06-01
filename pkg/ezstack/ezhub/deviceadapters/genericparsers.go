package deviceadapters

import (
	"github.com/function61/hautomo/pkg/ezstack"
	"github.com/function61/hautomo/pkg/ezstack/ezhub/hubtypes"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
)

// generic device with default parsers, used unless a specific device implements overridden parsers
var adapterByModel = map[ezstack.Model]Adapter{
	"*": newAdapter(
		attributeParser("genOnOff.onOff", genOnOffOnOff),
		attributeParser("msPressureMeasurement.measuredValue", msPressureMeasurementMeasuredValue),
		attributeParser("msPressureMeasurement.scaledValue", ignoreForNowTODOTakeIntoAccount),
		attributeParser("msPressureMeasurement.scale", ignoreForNowTODOTakeIntoAccount),
		attributeParser("msOccupancySensing.occupancy", msOccupancySensingOccupancy),
		attributeParser("msTemperatureMeasurement.measuredValue", msTemperatureMeasurementMeasuredValue),
		attributeParser("msRelativeHumidity.measuredValue", msRelativeHumidityMeasuredValue),
		attributeParser("msIlluminanceMeasurement.measuredValue", msIlluminanceMeasurementMeasuredValue),
		attributeParser("genBasic.modelId", noopParser), // we already got this key in our Zigbee device metadata, so don't record it
	),
}

func genOnOffOnOff(attr *cluster.Attribute, actx *hubtypes.AttrsCtx) error {
	actx.Attrs.On = actx.Bool(attr.Value.(bool))

	return nil
}

func msOccupancySensingOccupancy(attr *cluster.Attribute, actx *hubtypes.AttrsCtx) error {
	// TODO: zigbee2mqtt does modulo 2 and checks for odd - why?
	presence := attr.Value.(uint64) > 0

	actx.Attrs.Presence = actx.Bool(presence)

	return nil
}

func msPressureMeasurementMeasuredValue(attr *cluster.Attribute, actx *hubtypes.AttrsCtx) error {
	// FIXME: assuming measuredValue is in hPa
	actx.Attrs.Pressure = actx.Float(float64(attr.Value.(int64)))

	return nil
}

func msIlluminanceMeasurementMeasuredValue(attr *cluster.Attribute, actx *hubtypes.AttrsCtx) error {
	actx.Attrs.Illuminance = actx.Float(float64(attr.Value.(uint64)))

	return nil
}

func msTemperatureMeasurementMeasuredValue(attr *cluster.Attribute, actx *hubtypes.AttrsCtx) error {
	actx.Attrs.Temperature = actx.Float(float64(attr.Value.(int64)) / 100)

	return nil
}

func msRelativeHumidityMeasuredValue(attr *cluster.Attribute, actx *hubtypes.AttrsCtx) error {
	actx.Attrs.HumidityRelative = actx.Float(float64(attr.Value.(uint64)) / 100)

	return nil
}

func ignoreForNowTODOTakeIntoAccount(attr *cluster.Attribute, actx *hubtypes.AttrsCtx) error {
	return nil
}
