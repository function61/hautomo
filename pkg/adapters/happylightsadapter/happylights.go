package happylightsadapter

import (
	"github.com/function61/gokit/logger"
	"github.com/function61/home-automation-hub/hapitypes"
	"github.com/function61/home-automation-hub/libraries/happylights/client"
	"github.com/function61/home-automation-hub/libraries/happylights/types"
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
			select {
			case powerMsg := <-adapter.PowerMsg:
				bluetoothAddr := powerMsg.DeviceId

				var req types.LightRequest

				if powerMsg.On {
					req = types.LightRequestOn(bluetoothAddr)
				} else {
					req = types.LightRequestOff(bluetoothAddr)
				}

				if err := client.SendRequest(config.HappyLightsAddr, req); err != nil {
					log.Error(err.Error())
				}
			case brightnessMsg := <-adapter.BrightnessMsg:
				lastColor := brightnessMsg.LastColor
				brightness := brightnessMsg.Brightness

				dimmedColor := hapitypes.RGB{
					Red:   uint8(float64(lastColor.Red) * float64(brightness) / 100.0),
					Green: uint8(float64(lastColor.Green) * float64(brightness) / 100.0),
					Blue:  uint8(float64(lastColor.Blue) * float64(brightness) / 100.0),
				}

				// translate brightness directives into RGB directives
				handleColorMsg(hapitypes.NewColorMsg(brightnessMsg.DeviceId, dimmedColor))
			case colorMsg := <-adapter.ColorMsg:
				handleColorMsg(colorMsg)
			}
		}
	}()
}
