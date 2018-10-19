package harmonyhubadapter

import (
	"fmt"
	"github.com/function61/gokit/logger"
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/hapitypes"
	"github.com/function61/home-automation-hub/libraries/harmonyhub"
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
			case powerMsg := <-adapter.PowerMsg:
				if err := harmonyHubConnection.HoldAndRelease(powerMsg.DeviceId, powerMsg.PowerCommand); err != nil {
					log.Error(fmt.Sprintf("HoldAndRelease: %s", err.Error()))
				}
			case infraredMsg := <-adapter.InfraredMsg:
				if err := harmonyHubConnection.HoldAndRelease(infraredMsg.DeviceId, infraredMsg.Command); err != nil {
					log.Error(fmt.Sprintf("HoldAndRelease: %s", err.Error()))
				}
			}
		}
	}()
}
