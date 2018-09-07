package harmonyhubadapter

import (
	"github.com/function61/home-automation-hub/hapitypes"
	"github.com/function61/home-automation-hub/libraries/harmonyhub"
	"github.com/function61/home-automation-hub/util/stopper"
	"log"
)

func New(adapter *hapitypes.Adapter, config hapitypes.AdapterConfig, stop *stopper.Stopper) *hapitypes.Adapter {
	// because we don't own the given stopper, we shouldn't call Add() on it
	subStopper := stopper.New()

	harmonyHubConnection := harmonyhub.NewHarmonyHubConnection(config.HarmonyAddr, subStopper.Add())

	go func() {
		defer stop.Done()

		log.Println("HarmonyHubAdapter: started")

		for {
			select {
			case <-stop.ShouldStop:
				log.Println("HarmonyHubAdapter: stopping")
				subStopper.StopAll()
				return
			case powerMsg := <-adapter.PowerMsg:
				if err := harmonyHubConnection.HoldAndRelease(powerMsg.DeviceId, powerMsg.PowerCommand); err != nil {
					log.Printf("HarmonyHubAdapter: HoldAndRelease failed: %s", err.Error())
				}
			case infraredMsg := <-adapter.InfraredMsg:
				if err := harmonyHubConnection.HoldAndRelease(infraredMsg.DeviceId, infraredMsg.Command); err != nil {
					log.Printf("HarmonyHubAdapter: HoldAndRelease failed: %s", err.Error())
				}
			}
		}
	}()

	return adapter
}
