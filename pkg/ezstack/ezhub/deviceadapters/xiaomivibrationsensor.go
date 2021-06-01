package deviceadapters

import (
	"fmt"
	"math"

	"github.com/function61/hautomo/pkg/ezstack/ezhub/hubtypes"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
)

func init() {
	defineAdapter(modelAqaraVibrationSensor,
		aqaraVoltageEtc,
		withBatteryType(BatteryCR2032),
		attributeParser("closuresDoorLock.unknown(85)", aqaraVibrationAction),
		attributeParser("closuresDoorLock.unknown(1283)", aqaraVibrationAngle),
		attributeParser("closuresDoorLock.unknown(1285)", aqaraVibrationStrength),
		attributeParser("closuresDoorLock.unknown(1288)", aqaraVibrationOrientation),
	)
}

// https://github.com/Koenkk/zigbee-herdsman-converters/blob/f317842ed3c806cf1ed893a2b469def7cb361585/converters/fromZigbee.js#L5023

func aqaraVibrationAction(attr *cluster.Attribute, actx *hubtypes.AttrsCtx) error {
	val := attr.Value.(uint64)

	switch val {
	case 1:
		actx.Attrs.Vibration = actx.Event()
	case 2:
		actx.Attrs.Tilt = actx.Event()
	case 3:
		actx.Attrs.Drop = actx.Event()
	default:
		return fmt.Errorf("unknown action: %d", val)
	}

	return nil
}

func aqaraVibrationAngle(attr *cluster.Attribute, actx *hubtypes.AttrsCtx) error {
	angle := attr.Value.(uint64)

	// TODO: graduate to top-level attr
	actx.Attrs.CustomString["angle"] = actx.String(fmt.Sprintf("%d", angle))

	return nil
}

func aqaraVibrationStrength(attr *cluster.Attribute, actx *hubtypes.AttrsCtx) error {
	// "Only first 2 bytes are relevant."
	//   https://github.com/dresden-elektronik/deconz-rest-plugin/issues/748#issuecomment-419669995
	strength := attr.Value.(uint64) >> 8

	// Swap byte order. TODO: validate correct!
	strength = ((strength & 0xFF) << 8) | ((strength >> 8) & 0xFF)

	// TODO: graduate to top-level attr
	actx.Attrs.CustomString["strength"] = actx.String(fmt.Sprintf("%d", strength))

	return nil
}

func aqaraVibrationOrientation(attr *cluster.Attribute, actx *hubtypes.AttrsCtx) error {
	// 0491ffff0027
	// 6 bytes

	all := attr.Value.(uint64)

	// came in on the wire as Uint48. each component (from x/y/z) is 16 bits
	sixteenBits := uint64(0xFFFF)

	x := float64((all >> 0) & sixteenBits)
	y := float64((all >> 16) & sixteenBits)
	z := float64((all >> 32) & sixteenBits)

	// TODO: verify these
	actx.Attrs.Orientation = &hubtypes.AttrOrientation{
		X:          int(math.Round(math.Atan(x/math.Sqrt(y*y+z*z)) * 180 / math.Pi)),
		Y:          int(math.Round(math.Atan(y/math.Sqrt(x*x+z*z)) * 180 / math.Pi)),
		Z:          int(math.Round(math.Atan(z/math.Sqrt(x*x+y*y)) * 180 / math.Pi)),
		LastReport: actx.Reported,
	}

	// TODO: graduate to top-level attr
	// actx.Attrs.CustomString["angle2"] = actx.String(fmt.Sprintf("%x", all))
	// actx.Attrs.CustomString["angle3"] = actx.String(cluster.SerializeMagicValue(attr.DataType, attr.Value))

	return nil
}
