package harmonyhubadapter

import (
	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
	"github.com/function61/home-automation-hub/pkg/harmonyhub"
)

func Start(adapter *hapitypes.Adapter, stop *stopper.Stopper) error {
	// we cannot make hierarchical stoppers, but we can have "stop manager" inside a
	// stopper - it achieves the same thing
	stopManager := stopper.NewManager()

	harmonyhubEnableLogs := false

	harmonyhubLogger := logex.Prefix("lib", adapter.Log)
	if !harmonyhubEnableLogs {
		harmonyhubLogger = logex.Discard
	}

	harmonyHubConnection := harmonyhub.NewHarmonyHubConnection(
		adapter.Conf.HarmonyAddr,
		harmonyhubLogger,
		stopManager.Stopper())

	go func() {
		defer stop.Done()

		adapter.Logl.Info.Println("started")
		defer adapter.Logl.Info.Println("stopped")

		for {
			select {
			case <-stop.Signal:
				adapter.Logl.Info.Println("stopping")
				stopManager.StopAllWorkersAndWait()
				return
			case genericEvent := <-adapter.Outbound:
				switch e := genericEvent.(type) {
				case *hapitypes.PowerMsg:
					if err := harmonyHubConnection.HoldAndRelease(e.DeviceId, e.PowerCommand); err != nil {
						adapter.Logl.Error.Printf("HoldAndRelease: %s", err.Error())
					}
				case *hapitypes.InfraredEvent:
					if err := harmonyHubConnection.HoldAndRelease(e.Device, e.Command); err != nil {
						adapter.Logl.Error.Printf("HoldAndRelease: %s", err.Error())
					}
				default:
					adapter.LogUnsupportedEvent(genericEvent)
				}
			}
		}
	}()

	return nil
}
