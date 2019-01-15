package devicegroupadapter

import (
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
)

// this adapter just basically copies the outbound event as multiple copies with rewritten
// device ID and posts it as inbound again

func Start(adapter *hapitypes.Adapter, stop *stopper.Stopper) error {
	go func() {
		defer stop.Done()
		adapter.Logl.Info.Println("started")
		defer adapter.Logl.Info.Println("stopped")

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
	group := adapter.Conf.DevicegroupDevices // shorthand

	switch e := genericEvent.(type) {
	case *hapitypes.ColorMsg:
		for _, deviceId := range group {
			adapter.Receive(hapitypes.NewColorMsg(deviceId, e.Color))
		}
	case *hapitypes.PlaybackEvent:
		for _, deviceId := range group {
			adapter.Receive(hapitypes.NewPlaybackEvent(
				deviceId,
				e.Action))
		}
	case *hapitypes.BrightnessMsg:
		for _, deviceId := range group {
			adapter.Receive(hapitypes.NewBrightnessEvent(deviceId, e.Brightness))
		}
	case *hapitypes.ColorTemperatureEvent:
		for _, deviceId := range group {
			adapter.Receive(hapitypes.NewColorTemperatureEvent(
				deviceId,
				e.TemperatureInKelvin))
		}
	case *hapitypes.PowerMsg:
		for _, deviceId := range group {
			if e.On {
				adapter.Receive(hapitypes.NewPowerEvent(deviceId, hapitypes.PowerKindOn))
			} else {
				adapter.Receive(hapitypes.NewPowerEvent(deviceId, hapitypes.PowerKindOff))
			}
		}
	case *hapitypes.BlinkEvent:
		for _, deviceId := range group {
			adapter.Receive(hapitypes.NewBlinkEvent(deviceId))
		}
	default:
		adapter.LogUnsupportedEvent(genericEvent)
	}
}
