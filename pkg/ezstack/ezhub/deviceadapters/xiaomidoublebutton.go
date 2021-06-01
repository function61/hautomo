package deviceadapters

import (
	"fmt"

	"github.com/function61/hautomo/pkg/evdevcodes"
	"github.com/function61/hautomo/pkg/ezstack/ezhub/hubtypes"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
	"github.com/function61/hautomo/pkg/ezstack/zigbee"
)

func init() {
	leftButton := evdevcodes.Btn0
	rightButton := evdevcodes.Btn1

	/*	This device uses Zigbee endpoints to model different buttons / combinations

		1 => left button
		2 => right button
		3 => both

		We'll translate these to more semantic values. Endpoint 3 to mean "both buttons" is not a
		pretty design. If you had three buttons, how many permutations would "two buttons pressed" need?
		How about 9 buttons?
	*/
	endpointToPressedButtons := map[zigbee.EndpointId][]evdevcodes.KeyOrButton{
		1: {leftButton},
		2: {rightButton},
		3: {leftButton, rightButton},
	}

	defineAdapter(modelAqaraDoubleButtonSensor,
		aqaraVoltageEtc,
		withBatteryType(BatteryCR2032),
		attributeParser("genMultistateInput.presentValue", func(attr *cluster.Attribute, actx *hubtypes.AttrsCtx) error {
			pressCount := int(attr.Value.(uint64))
			isHold := pressCount == 0

			pressedButtons, found := endpointToPressedButtons[actx.Endpoint]
			if !found {
				return fmt.Errorf("unexpected endpoint: %d", actx.Endpoint)
			}

			actx.Attrs.Press = &hubtypes.AttrPress{
				Key:            pressedButtons[0],
				KeysAdditional: pressedButtons[1:],
				Kind: func() hubtypes.PressKind {
					if isHold {
						return hubtypes.PressKindHold
					} else {
						return hubtypes.PressKindUp
					}
				}(),
				CountRaw: func() *int {
					if pressCount >= 2 {
						return &pressCount
					} else {
						return nil
					}
				}(),

				LastReport: actx.Reported,
			}

			return nil
		}),
	)
}
