package ikeatradfriadapter

import (
	"fmt"
	"github.com/function61/gokit/logger"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
	"github.com/function61/home-automation-hub/pkg/ikeatradfri"
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
			genericEvent := <-adapter.Event

			switch e := genericEvent.(type) {
			case *hapitypes.PowerMsg:
				var responseErr error = nil

				if e.On {
					responseErr = ikeatradfri.TurnOn(e.DeviceId, coapClient)
				} else {
					responseErr = ikeatradfri.TurnOff(e.DeviceId, coapClient)
				}

				if responseErr != nil {
					log.Error(responseErr.Error())
				}
			case *hapitypes.BrightnessMsg:
				// 0-100 => 0-254
				to := int(float64(e.Brightness) * 2.54)

				if err := ikeatradfri.Dim(e.DeviceId, to, coapClient); err != nil {
					log.Error(fmt.Sprintf("Dim: %s", err.Error()))
				}
			case *hapitypes.ColorMsg:
				if err := ikeatradfri.SetRGB(e.DeviceId, e.Color.Red, e.Color.Green, e.Color.Blue, coapClient); err != nil {
					log.Error(err.Error())
				}
			case *hapitypes.ColorTemperatureEvent:
				if err := ikeatradfri.SetColorTemp(
					e.Device,
					e.TemperatureInKelvin,
					coapClient); err != nil {
					log.Error(err.Error())
				}
			default:
				adapter.LogUnsupportedEvent(genericEvent, log)
			}
		}
	}()
}
