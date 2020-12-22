package ikeatradfriadapter

import (
	"context"

	"github.com/function61/hautomo/pkg/hapitypes"
	"github.com/function61/hautomo/pkg/ikeatradfri"
)

func Start(ctx context.Context, adapter *hapitypes.Adapter) error {
	coapClient := ikeatradfri.NewCoapClient(
		adapter.Conf.Url,
		adapter.Conf.TradfriUser,
		adapter.Conf.TradfriPsk)

	for {
		select {
		case <-ctx.Done():
			return nil
		case genericEvent := <-adapter.Outbound:
			handleEvent(genericEvent, coapClient, adapter)
		}
	}
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
			adapter.Logl.Error.Println(responseErr.Error())
		}
	case *hapitypes.BrightnessMsg:
		// 0-100 => 0-254
		to := int(float64(e.Brightness) * 2.54)

		if err := ikeatradfri.Dim(e.DeviceId, to, coapClient); err != nil {
			adapter.Logl.Error.Printf("Dim: %s", err.Error())
		}
	case *hapitypes.ColorMsg:
		if err := ikeatradfri.SetRGB(e.DeviceId, e.Color.Red, e.Color.Green, e.Color.Blue, coapClient); err != nil {
			adapter.Logl.Error.Println(err.Error())
		}
	case *hapitypes.ColorTemperatureEvent:
		if err := ikeatradfri.SetColorTemp(
			e.Device,
			e.TemperatureInKelvin,
			coapClient); err != nil {
			adapter.Logl.Error.Println(err.Error())
		}
	default:
		adapter.LogUnsupportedEvent(genericEvent)
	}
}
