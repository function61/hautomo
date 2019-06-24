package sonoffadapter

import (
	"context"
	"github.com/function61/gokit/stopper"
	"github.com/function61/hautomo/pkg/hapitypes"
	"github.com/function61/hautomo/pkg/sonoff"
	"time"
)

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
	switch e := genericEvent.(type) {
	case *hapitypes.PowerMsg:
		ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancel()

		var err error

		if e.On {
			err = sonoff.TurnOn(ctx, e.DeviceId)
		} else {
			err = sonoff.TurnOff(ctx, e.DeviceId)
		}

		if err != nil {
			adapter.Logl.Error.Println(err.Error())
		}
	default:
		adapter.LogUnsupportedEvent(genericEvent)
	}
}
