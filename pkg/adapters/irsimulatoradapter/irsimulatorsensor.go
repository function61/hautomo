package irsimulatoradapter

import (
	"github.com/function61/gokit/logger"
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
	"time"
)

var log = logger.New("IR simulator")

func Start(adapter *hapitypes.Adapter, stop *stopper.Stopper) error {
	go func() {
		defer stop.Done()

		log.Info("started")
		defer log.Info("stopping")

		for {
			select {
			case <-stop.Signal:
				return
			case <-time.After(5 * time.Second):
				e := hapitypes.NewInfraredEvent("simulated_remote", adapter.Conf.IrSimulatorKey)
				adapter.Inbound.Receive(&e)
			}
		}
	}()

	return nil
}
