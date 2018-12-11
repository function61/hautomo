package particleadapter

import (
	"errors"
	"github.com/function61/gokit/logger"
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
	"github.com/function61/home-automation-hub/pkg/particleapi"
)

var log = logger.New("particleadapter")

func Start(adapter *hapitypes.Adapter, stop *stopper.Stopper) error {
	if adapter.Conf.ParticleAccessToken == "" || adapter.Conf.ParticleId == "" {
		return errors.New("ParticleAccessToken or ParticleId not defined")
	}

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
		if err := particleapi.Invoke(adapter.Conf.ParticleId, "rf", e.PowerCommand, adapter.Conf.ParticleAccessToken); err != nil {
			log.Error(err.Error())
		}
	default:
		adapter.LogUnsupportedEvent(genericEvent, log)
	}
}
