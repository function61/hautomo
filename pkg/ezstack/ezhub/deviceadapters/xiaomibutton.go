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
		attributeParser("genOnOff.onOff", xiaomiButtonGenOnOffOnOff),
		attributeParser("genOnOff.unknown(32768)", xiaomiButtonMultiClick),
	)
}

func xiaomiButtonGenOnOffOnOff(attr *cluster.Attribute, actx *hubtypes.AttrsCtx) error {
	// Xiaomi push button sends single click as two events:
	// power=off
	// power=on

	// only react on "on", so we don't broadcast multiple state changes.
	// they might come in a single Zigbee message which makes this a non-issue, but better safe.
	if attr.Value.(bool) {
		actx.Attrs.Press = actx.PressUp(evdevcodes.Btn0)
	}

	return nil
}

// it sends click >= 2 as manufacturer-specific attribute
func xiaomiButtonMultiClick(attr *cluster.Attribute, actx *hubtypes.AttrsCtx) error {
	count := int(attr.Value.(uint64))

	actx.Attrs.Press = actx.PressUp(evdevcodes.Btn0)
	if count >= 2 { // FIXME: dirty setting after-the-fact
		actx.Attrs.Press.CountRaw = &count
	}

	return nil
}
