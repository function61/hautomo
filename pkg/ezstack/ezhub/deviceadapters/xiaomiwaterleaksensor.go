package deviceadapters

import (
	"fmt"

	"github.com/function61/hautomo/pkg/ezstack/ezhub/hubtypes"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
)

func init() {
	defineAdapter(modelAqaraWaterLeakSensor,
		aqaraVoltageEtc,
		withBatteryType(BatteryCR2032),
		withCommandHandler(func(command interface{}, actx *hubtypes.AttrsCtx) error {
			switch cmd := command.(type) {
			case *cluster.ZoneStatusChangeNotificationCommand:
				if cmd.ZoneStatus.UnimplementedBitsSet() {
					// FIXME: log warning instead?
					return fmt.Errorf("ZoneStatus has unimplemented bits set: %b", cmd.ZoneStatus)
				} else {
					actx.Attrs.WaterDetected = actx.Bool(cmd.ZoneStatus.Alarm1())
				}

				return nil
			default:
				return errUnhandledCommand
			}
		}))
}
