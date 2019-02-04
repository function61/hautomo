package trionesadapter

import (
	"context"
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
	"github.com/function61/home-automation-hub/pkg/triones"
	"log"
	"time"
)

const requestTimeout = 15 * time.Second

func Start(adapter *hapitypes.Adapter, stop *stopper.Stopper) error {
	conf := adapter.GetConfigFileDeprecated()

	go func() {
		defer stop.Done()

		adapter.Logl.Info.Println("started")
		defer adapter.Logl.Info.Println("stopped")

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

		var req triones.Request

		if e.On {
			req = triones.RequestOn(bluetoothAddr)
		} else {
			req = triones.RequestOff(bluetoothAddr)
		}

		if err := sendLightRequest(req, adapter.Log); err != nil {
			adapter.Logl.Error.Println(err.Error())
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
		adapter.Send(hapitypes.NewColorMsg(e.DeviceId, dimmedColor))
	case *hapitypes.ColorMsg:
		bluetoothAddr := e.DeviceId

		deviceConf := conf.FindDeviceConfigByAdaptersDeviceId(bluetoothAddr)
		deviceType, err := hapitypes.ResolveDeviceType(deviceConf.Type)
		if err != nil {
			panic(err)
		}
		caps := deviceType.Capabilities

		var req triones.Request
		if e.Color.IsGrayscale() && caps.ColorSeparateWhiteChannel {
			// we can just take red because we know that r == g == b
			req = triones.RequestWhite(bluetoothAddr, e.Color.Red)
		} else {
			req = triones.RequestRGB(
				bluetoothAddr,
				e.Color.Red,
				e.Color.Green,
				e.Color.Blue)

			// I don't know if my only Triones controller that is attached to a RGBW strip
			// is messed up, or if the pinouts of this controller and this particular strip
			// that are incompatible, but here Red and Green channels are mixed up.
			// compensating for it here.
			if caps.ColorSeparateWhiteChannel {
				// swap red <-> green channels
				temp := req.RgbOpts.Red
				req.RgbOpts.Red = req.RgbOpts.Green
				req.RgbOpts.Green = temp
			}
		}

		if err := sendLightRequest(req, adapter.Log); err != nil {
			adapter.Logl.Error.Println(err.Error())
		}
	default:
		adapter.LogUnsupportedEvent(genericEvent)
	}
}

func sendLightRequest(hlreq triones.Request, logger *log.Logger) error {
	ctx, cancel := context.WithTimeout(context.TODO(), requestTimeout)
	defer cancel()

	return triones.Send(ctx, hlreq, logger)
}
