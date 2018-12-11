package devicegroupadapter

import (
	"github.com/function61/gokit/logger"
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
)

var log = logger.New("devicegroupadapter")

// this adapter just basically copies the outbound event as multiple copies with rewritten
// device ID and posts it as inbound again

func Start(adapter *hapitypes.Adapter, stop *stopper.Stopper) error {
	go func() {
		defer stop.Done()
		log.Info("started")
		defer log.Info("stopped")

		for {
			select {
			case <-stop.Signal:
				return
			case genericEvent := <-adapter.Outbound:
				handleEvent(genericEvent, adapter)
			}
		}
	}()

	return nil
}

func handleEvent(genericEvent hapitypes.OutboundEvent, adapter *hapitypes.Adapter) {
	inbound := adapter.Inbound // shorthands
	conf := adapter.Conf

	switch e := genericEvent.(type) {
	case *hapitypes.ColorMsg:
		for _, deviceId := range conf.DevicegroupDevices {
			e2 := hapitypes.NewColorMsg(deviceId, e.Color)
			inbound.Receive(&e2)
		}
	case *hapitypes.PlaybackEvent:
		for _, deviceId := range conf.DevicegroupDevices {
			e2 := hapitypes.NewPlaybackEvent(deviceId, e.Action)
			inbound.Receive(&e2)
		}
	case *hapitypes.BrightnessMsg:
		for _, deviceId := range conf.DevicegroupDevices {
			e2 := hapitypes.NewBrightnessEvent(deviceId, e.Brightness)
			inbound.Receive(&e2)
		}
	case *hapitypes.ColorTemperatureEvent:
		for _, deviceId := range conf.DevicegroupDevices {
			e2 := hapitypes.NewColorTemperatureEvent(deviceId, e.TemperatureInKelvin)
			inbound.Receive(&e2)
		}
	case *hapitypes.PowerMsg:
		for _, deviceId := range conf.DevicegroupDevices {
			var e2 hapitypes.PowerEvent
			if e.On {
				e2 = hapitypes.NewPowerEvent(deviceId, hapitypes.PowerKindOn)
			} else {
				e2 = hapitypes.NewPowerEvent(deviceId, hapitypes.PowerKindOff)
			}

			inbound.Receive(&e2)
		}
	default:
		adapter.LogUnsupportedEvent(genericEvent, log)
	}
}
