package happylightsadapter

import (
	"github.com/function61/gokit/logger"
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
	"github.com/function61/home-automation-hub/pkg/happylights/client"
	"github.com/function61/home-automation-hub/pkg/happylights/types"
)

var log = logger.New("HappyLights")

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
	switch e := genericEvent.(type) {
	case *hapitypes.PowerMsg:
		bluetoothAddr := e.DeviceId

		var req types.LightRequest

		if e.On {
			req = types.LightRequestOn(bluetoothAddr)
		} else {
			req = types.LightRequestOff(bluetoothAddr)
		}

		if err := client.SendRequest(adapter.Conf.HappyLightsAddr, req); err != nil {
			log.Error(err.Error())
		}
	case *hapitypes.BrightnessMsg:
		lastColor := e.LastColor
		brightness := e.Brightness

		dimmedColor := hapitypes.NewRGB(
			uint8(float64(lastColor.Red)*float64(brightness)/100.0),
			uint8(float64(lastColor.Green)*float64(brightness)/100.0),
			uint8(float64(lastColor.Blue)*float64(brightness)/100.0),
		)

		// translate brightness directives into RGB directives
		handleColorMsg(hapitypes.NewColorMsg(e.DeviceId, dimmedColor), adapter)
	case *hapitypes.ColorMsg:
		handleColorMsg(*e, adapter)
	default:
		adapter.LogUnsupportedEvent(genericEvent, log)
	}
}

func handleColorMsg(colorMsg hapitypes.ColorMsg, adapter *hapitypes.Adapter) {
	bluetoothAddr := colorMsg.DeviceId

	hlreq := types.LightRequestColor(
		bluetoothAddr,
		colorMsg.Color.Red,
		colorMsg.Color.Green,
		colorMsg.Color.Blue)

	if err := client.SendRequest(adapter.Conf.HappyLightsAddr, hlreq); err != nil {
		log.Error(err.Error())
	}
}
