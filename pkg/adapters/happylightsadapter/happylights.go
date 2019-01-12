package happylightsadapter

import (
	"context"
	"github.com/function61/gokit/logger"
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
	"github.com/function61/home-automation-hub/pkg/happylights"
	"time"
)

var log = logger.New("HappyLights")

const requestTimeout = 15 * time.Second

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

		var req happylights.LightRequest

		if e.On {
			req = happylights.LightRequestOn(bluetoothAddr)
		} else {
			req = happylights.LightRequestOff(bluetoothAddr)
		}

		sendLightRequest(req)
	case *hapitypes.BrightnessMsg:
		lastColor := e.LastColor
		brightness := e.Brightness

		dimmedColor := hapitypes.NewRGB(
			uint8(float64(lastColor.Red)*float64(brightness)/100.0),
			uint8(float64(lastColor.Green)*float64(brightness)/100.0),
			uint8(float64(lastColor.Blue)*float64(brightness)/100.0),
		)

		// translate brightness directives into RGB directives
		colorMsg := hapitypes.NewColorMsg(e.DeviceId, dimmedColor)
		adapter.Send(&colorMsg)
	case *hapitypes.ColorMsg:
		bluetoothAddr := e.DeviceId

		sendLightRequest(happylights.LightRequestColor(
			bluetoothAddr,
			e.Color.Red,
			e.Color.Green,
			e.Color.Blue))
	default:
		adapter.LogUnsupportedEvent(genericEvent, log)
	}
}

func sendLightRequest(hlreq happylights.LightRequest) {
	ctx, cancel := context.WithTimeout(context.TODO(), requestTimeout)
	defer cancel()

	if err := happylights.Send(ctx, hlreq); err != nil {
		log.Error(err.Error())
	}
}
