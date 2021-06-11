package deviceadapters

import (
	"github.com/function61/hautomo/pkg/evdevcodes"
	"github.com/function61/hautomo/pkg/ezstack/ezhub/hubtypes"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
)

func init() {
	// see E1524 remote for explanation of these key mappings

	defineAdapter("TRADFRI open/close remote",
		withBatteryType(BatteryCR2032),
		withCommandHandler(func(command interface{}, actx *hubtypes.AttrsCtx) error {
			switch command.(type) {
			case *cluster.ClosuresWindowCoveringUp:
				actx.Attrs.Press = actx.PressUp(evdevcodes.KeyOPEN)

				return nil
			case *cluster.ClosuresWindowCoveringDown:
				actx.Attrs.Press = actx.PressUp(evdevcodes.KeyCLOSE)

				return nil
			default:
				return errUnhandledCommand
			}
		}))
}
