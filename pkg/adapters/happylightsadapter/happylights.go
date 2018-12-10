package happylightsadapter

import (
	"github.com/function61/gokit/logger"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
	"github.com/function61/home-automation-hub/pkg/happylights/client"
	"github.com/function61/home-automation-hub/pkg/happylights/types"
)

var log = logger.New("HappyLights")

func New(adapter *hapitypes.Adapter, config hapitypes.AdapterConfig) {
	handleColorMsg := func(colorMsg hapitypes.ColorMsg) {
		bluetoothAddr := colorMsg.DeviceId

		hlreq := types.LightRequestColor(
			bluetoothAddr,
			colorMsg.Color.Red,
			colorMsg.Color.Green,
			colorMsg.Color.Blue)

		if err := client.SendRequest(config.HappyLightsAddr, hlreq); err != nil {
			log.Error(err.Error())
		}
	}

	go func() {
		log.Info("started")

		for {
			genericEvent := <-adapter.Event

			switch e := genericEvent.(type) {
			case *hapitypes.PowerMsg:
				bluetoothAddr := e.DeviceId

				var req types.LightRequest

				if e.On {
					req = types.LightRequestOn(bluetoothAddr)
				} else {
					req = types.LightRequestOff(bluetoothAddr)
				}

				if err := client.SendRequest(config.HappyLightsAddr, req); err != nil {
					log.Error(err.Error())
				}
			case *hapitypes.BrightnessMsg:
				lastColor := e.LastColor
				brightness := e.Brightness

				dimmedColor := hapitypes.RGB{
					Red:   uint8(float64(lastColor.Red) * float64(brightness) / 100.0),
					Green: uint8(float64(lastColor.Green) * float64(brightness) / 100.0),
					Blue:  uint8(float64(lastColor.Blue) * float64(brightness) / 100.0),
				}

				// translate brightness directives into RGB directives
				handleColorMsg(hapitypes.NewColorMsg(e.DeviceId, dimmedColor))
			case *hapitypes.ColorMsg:
				handleColorMsg(*e)
			default:
				adapter.LogUnsupportedEvent(genericEvent, log)
			}
		}
	}()
}
