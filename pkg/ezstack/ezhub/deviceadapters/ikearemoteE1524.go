package deviceadapters

import (
	"github.com/function61/hautomo/pkg/evdevcodes"
	"github.com/function61/hautomo/pkg/ezstack/ezhub/hubtypes"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
)

func init() {
	// it's sad that Zigbee remotes generally don't send their keypresses, but instead try to control
	// concrete end-device (like a light). so the remotes are already paired to a certain class of
	// device (light remote, shade remote, etc.), so when our intent is to only detect the buttons
	// pressed we've to hackily translate "move to previous scene" command back to a "left key".
	defineAdapter("TRADFRI remote control",
		withBatteryType(BatteryCR2032),
		withCommandHandler(func(command interface{}, actx *hubtypes.AttrsCtx) error {
			switch cmd := command.(type) {
			case *cluster.GenOnOffToggleCommand:
				actx.Attrs.Press = actx.PressUp(evdevcodes.KeyPOWER)

				return nil
			case *cluster.StepOnOffCommand:
				actx.Attrs.Press = func() *hubtypes.AttrPress {
					if cmd.StepMode == 0 { // up
						return actx.PressUp(evdevcodes.KeyBRIGHTNESSUP)
					} else {
						return actx.PressUp(evdevcodes.KeyBRIGHTNESSDOWN)
					}
				}()

				return nil
			case *cluster.StepCommand:
				actx.Attrs.Press = func() *hubtypes.AttrPress {
					if cmd.StepMode == 0 { // up
						return actx.PressUp(evdevcodes.KeyBRIGHTNESSUP)
					} else {
						return actx.PressUp(evdevcodes.KeyBRIGHTNESSDOWN)
					}
				}()

				return nil
			case *cluster.ScenesMysteryCommand7:
				actx.Attrs.Press = func() *hubtypes.AttrPress {
					if cmd.Left() {
						return actx.PressUp(evdevcodes.KeyLEFT)
					} else {
						return actx.PressUp(evdevcodes.KeyRIGHT)
					}
				}()

				return nil
			default:
				return errUnhandledCommand
			}
		}))
}
