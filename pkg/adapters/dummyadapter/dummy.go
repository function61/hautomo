package dummyadapter

// dummy adapter just discards and logs messages

import (
	"github.com/function61/gokit/stopper"
	"github.com/function61/hautomo/pkg/hapitypes"
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
			case e := <-adapter.Outbound:
				adapter.LogUnsupportedEvent(e)
			}
		}
	}()

	return nil
}
