package ikeatradfriadapter

import (
	"fmt"
	"github.com/function61/gokit/logger"
	"github.com/function61/home-automation-hub/hapitypes"
	"github.com/function61/home-automation-hub/libraries/ikeatradfri"
)

var log = logger.New("ikeatradfriadapter")

func New(adapter *hapitypes.Adapter, config hapitypes.AdapterConfig) {
	go func() {
		log.Info("started")
		defer log.Info("stopped")

		coapClient := ikeatradfri.NewCoapClient(
			config.TradfriUrl,
			config.TradfriUser,
			config.TradfriPsk)

		for {
			select {
			case powerMsg := <-adapter.PowerMsg:
				var responseErr error = nil

				if powerMsg.On {
					responseErr = ikeatradfri.TurnOn(powerMsg.DeviceId, coapClient)
				} else {
					responseErr = ikeatradfri.TurnOff(powerMsg.DeviceId, coapClient)
				}

				if responseErr != nil {
					log.Error(responseErr.Error())
				}
			case brightnessMsg := <-adapter.BrightnessMsg:
				// 0-100 => 0-254
				to := int(float64(brightnessMsg.Brightness) * 2.54)

				if err := ikeatradfri.DimWithoutFading(brightnessMsg.DeviceId, to, coapClient); err != nil {
					log.Error(fmt.Sprintf("DimWithoutFading: %s", err.Error()))
				}
			}
		}
	}()
}
