package deviceadapters

import (
	"github.com/function61/hautomo/pkg/ezstack/ezhub/hubtypes"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
)

func init() {
	defineAdapter(modelAqaraDoorSensor,
		aqaraVoltageEtc,
		withBatteryType(BatteryCR2032),
		attributeParser("genOnOff.onOff", aqaraDoorSensorGenOnOffOnOff),
	)
}

func aqaraDoorSensorGenOnOffOnOff(attr *cluster.Attribute, actx *hubtypes.AttrsCtx) error {
	actx.Attrs.Contact = actx.Bool(attr.Value.(bool))

	return nil
}
