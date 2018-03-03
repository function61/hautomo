package main

import (
	"./hapitypes"
	"log"
)

func NewHarmonyHubAdapter(id string, addr string, stopper *Stopper) *hapitypes.Adapter {
	adapter := hapitypes.NewAdapter(id)

	// because we don't own the given stopper, we shouldn't call Add() on it
	subStopper := NewStopper()

	harmonyHubConnection := NewHarmonyHubConnection(addr, subStopper.Add())

	go func() {
		defer stopper.Done()

		log.Println("HarmonyHubAdapter: started")

		for {
			select {
			case <-stopper.ShouldStop:
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
