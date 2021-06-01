package deviceadapters

import (
	"github.com/function61/hautomo/pkg/evdevcodes"
	"github.com/function61/hautomo/pkg/ezstack/ezhub/hubtypes"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
)

func init() {
	defineAdapter(modelAqaraButtonSensor,
		aqaraVoltageEtc,
		withBatteryType(BatteryCR2032),
		attributeParser("genOnOff.onOff", aqaraButtonGenOnOffOnOff),
		attributeParser("genOnOff.unknown(32768)", aqaraButtonMultiClick),
	)
}

// Aqara push button sends single click as two events:
// power=off
// power=on
func aqaraButtonGenOnOffOnOff(attr *cluster.Attribute, actx *hubtypes.AttrsCtx) error {
	actx.Attrs.Press = actx.PressUp(evdevcodes.Btn0)

	return nil
}

// it sends click >= 2 as manufacturer-specific attribute
func aqaraButtonMultiClick(attr *cluster.Attribute, actx *hubtypes.AttrsCtx) error {
	count := int(attr.Value.(uint64))

	actx.Attrs.Press = actx.PressUp(evdevcodes.Btn0)
	if count >= 2 { // FIXME: dirty setting after-the-fact
		actx.Attrs.Press.CountRaw = &count
	}

	return nil
}
