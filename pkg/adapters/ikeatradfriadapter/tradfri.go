package ikeatradfriadapter

import (
	"fmt"
	"github.com/function61/gokit/logger"
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
	"github.com/function61/home-automation-hub/pkg/ikeatradfri"
)

var log = logger.New("ikeatradfriadapter")

func Start(adapter *hapitypes.Adapter, stop *stopper.Stopper) error {
	coapClient := ikeatradfri.NewCoapClient(
		adapter.Conf.TradfriUrl,
		adapter.Conf.TradfriUser,
		adapter.Conf.TradfriPsk)

	go func() {
		defer stop.Done()

		log.Info("started")
		defer log.Info("stopped")

		for {
			select {
			case <-stop.Signal:
				return
			case genericEvent := <-adapter.Outbound:
				handleEvent(genericEvent, coapClient, adapter)
			}
		}
	}()

	return nil
}

func handleEvent(genericEvent hapitypes.OutboundEvent, coapClient *ikeatradfri.CoapClient, adapter *hapitypes.Adapter) {
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
