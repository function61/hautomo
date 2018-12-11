package devicegroupadapter

import (
	"github.com/function61/gokit/logger"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
	"github.com/function61/home-automation-hub/pkg/signalfabric"
)

var log = logger.New("devicegroupadapter")

// this adapter just basically copies the outbound event as multiple copies with rewritten
// device ID and posts it as inbound again

func New(adapter *hapitypes.Adapter, inbound *signalfabric.Fabric, config hapitypes.AdapterConfig) {
	go func() {
		log.Info("started")

		for {
			genericEvent := <-adapter.Event

			switch e := genericEvent.(type) {
			case *hapitypes.ColorMsg:
				for _, deviceId := range config.DevicegroupDevices {
					e2 := hapitypes.NewColorMsg(deviceId, e.Color)
					inbound.Receive(&e2)
				}
			case *hapitypes.PlaybackEvent:
				for _, deviceId := range config.DevicegroupDevices {
					e2 := hapitypes.NewPlaybackEvent(deviceId, e.Action)
					inbound.Receive(&e2)
				}
			case *hapitypes.BrightnessMsg:
				for _, deviceId := range config.DevicegroupDevices {
					e2 := hapitypes.NewBrightnessEvent(deviceId, e.Brightness)
					inbound.Receive(&e2)
				}
			case *hapitypes.ColorTemperatureEvent:
				for _, deviceId := range config.DevicegroupDevices {
					e2 := hapitypes.NewColorTemperatureEvent(deviceId, e.TemperatureInKelvin)
					inbound.Receive(&e2)
				}
			case *hapitypes.PowerMsg:
				for _, deviceId := range config.DevicegroupDevices {
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
	}()
}
