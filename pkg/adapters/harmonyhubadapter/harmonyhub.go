package harmonyhubadapter

import (
	"fmt"
	"github.com/function61/gokit/logger"
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
	"github.com/function61/home-automation-hub/pkg/harmonyhub"
)

var log = logger.New("HarmonyHubAdapter")

func New(adapter *hapitypes.Adapter, config hapitypes.AdapterConfig, stop *stopper.Stopper) {
	// we cannot make hierarchical stoppers, but we can have "stop manager" inside a
	// stopper - it achieves the same thing
	stopManager := stopper.NewManager()

	harmonyHubConnection := harmonyhub.NewHarmonyHubConnection(config.HarmonyAddr, stopManager.Stopper())

	go func() {
		defer stop.Done()

		log.Info("started")
		defer log.Info("stopped")

		for {
			select {
			case <-stop.Signal:
				log.Info("stopping")
				stopManager.StopAllWorkersAndWait()
				return
			case genericEvent := <-adapter.Event:
				switch e := genericEvent.(type) {
				case *hapitypes.PowerMsg:
					if err := harmonyHubConnection.HoldAndRelease(e.DeviceId, e.PowerCommand); err != nil {
						log.Error(fmt.Sprintf("HoldAndRelease: %s", err.Error()))
					}
				case *hapitypes.InfraredMsg:
					if err := harmonyHubConnection.HoldAndRelease(e.DeviceId, e.Command); err != nil {
						log.Error(fmt.Sprintf("HoldAndRelease: %s", err.Error()))
					}
				default:
					adapter.LogUnsupportedEvent(genericEvent, log)
				}
			}
		}
	}()
}
