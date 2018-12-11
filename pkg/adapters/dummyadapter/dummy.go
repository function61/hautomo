package dummyadapter

// dummy adapter just discards and logs messages

import (
	"github.com/function61/gokit/logger"
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
)

var log = logger.New("dummyadapter")

func Start(adapter *hapitypes.Adapter, stop *stopper.Stopper) error {
	go func() {
		defer stop.Done()
		log.Info("started")
		defer log.Info("stopped")

		for {
			select {
			case <-stop.Signal:
				return
			case e := <-adapter.Outbound:
				adapter.LogUnsupportedEvent(e, log)
			}
		}
	}()

	return nil
}
