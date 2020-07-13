package irsimulatoradapter

import (
	"time"

	"github.com/function61/gokit/stopper"
	"github.com/function61/hautomo/pkg/hapitypes"
)

func Start(adapter *hapitypes.Adapter, stop *stopper.Stopper) error {
	go func() {
		defer stop.Done()

		adapter.Logl.Info.Println("started")
		defer adapter.Logl.Info.Println("stopping")

		for {
			select {
			case <-stop.Signal:
				return
			case <-time.After(5 * time.Second):
				adapter.Receive(hapitypes.NewRawInfraredEvent(
					"simulated_remote",
					adapter.Conf.IrSimulatorKey))
			}
		}
	}()

	return nil
}
