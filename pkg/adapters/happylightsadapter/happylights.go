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
	conf := adapter.GetConfigFileDeprecated()

	go func() {
		defer stop.Done()

		log.Info("started")
		defer log.Info("stopped")

		for {
			select {
			case <-stop.Signal:
				return
			case genericEvent := <-adapter.Outbound:
				handleEvent(genericEvent, adapter, conf)
			}
		}
	}()

	return nil
}

func handleEvent(genericEvent hapitypes.OutboundEvent, adapter *hapitypes.Adapter, conf *hapitypes.ConfigFile) {
	switch e := genericEvent.(type) {
	case *hapitypes.PowerMsg:
		bluetoothAddr := e.DeviceId

		var req happylights.Request

		if e.On {
			req = happylights.RequestOn(bluetoothAddr)
		} else {
			req = happylights.RequestOff(bluetoothAddr)
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

		deviceConf := conf.FindDeviceConfigByAdaptersDeviceId(bluetoothAddr)

		var req happylights.Request
		if e.Color.IsGrayscale() && deviceConf.CapabilityColorSeparateWhiteChannel {
			// we can just take red because we know that r == g == b
			req = happylights.RequestWhite(bluetoothAddr, e.Color.Red)
		} else {
			req = happylights.RequestRGB(
				bluetoothAddr,
				e.Color.Red,
				e.Color.Green,
				e.Color.Blue)
		}

		sendLightRequest(req)
	default:
		adapter.LogUnsupportedEvent(genericEvent, log)
	}
}

func sendLightRequest(hlreq happylights.Request) {
	ctx, cancel := context.WithTimeout(context.TODO(), requestTimeout)
	defer cancel()

	if err := happylights.Send(ctx, hlreq); err != nil {
		log.Error(err.Error())
	}
}
