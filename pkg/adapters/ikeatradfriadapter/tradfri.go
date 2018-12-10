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

				if err := ikeatradfri.DimWithoutFading(e.DeviceId, to, coapClient); err != nil {
					log.Error(fmt.Sprintf("DimWithoutFading: %s", err.Error()))
				}
			case *hapitypes.ColorTemperatureEvent:
				if err := ikeatradfri.SetColorTemp(
					e.Device,
					tempFromKelvin(e.TemperatureInKelvin),
					coapClient); err != nil {
					log.Error(err.Error())
				}
			default:
				adapter.LogUnsupportedEvent(genericEvent, log)
			}
		}
	}()
}

func tempFromKelvin(kelvin uint) ikeatradfri.ColorTemp {
	// https://developer.amazon.com/docs/device-apis/alexa-colortemperaturecontroller.html#setcolortemperature
	if kelvin < 4000 {
		return ikeatradfri.ColorTempWarm
	}

	if kelvin < 7000 {
		return ikeatradfri.ColorTempNormal
	}

	return ikeatradfri.ColorTempCold
}
