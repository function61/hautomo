package particleadapter

import (
	"errors"

	"github.com/function61/gokit/stopper"
	"github.com/function61/hautomo/pkg/hapitypes"
	"github.com/function61/hautomo/pkg/particleapi"
)

func Start(adapter *hapitypes.Adapter, stop *stopper.Stopper) error {
	if adapter.Conf.ParticleAccessToken == "" || adapter.Conf.ParticleId == "" {
		return errors.New("ParticleAccessToken or ParticleId not defined")
	}

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
		if err := particleapi.Invoke(adapter.Conf.ParticleId, "rf", e.PowerCommand, adapter.Conf.ParticleAccessToken); err != nil {
			adapter.Logl.Error.Println(err.Error())
		}
	default:
		adapter.LogUnsupportedEvent(genericEvent)
	}
}
