package main

import (
	"log"
)

func NewHarmonyHubAdapter(id string, addr string, stopper *Stopper) *Adapter {
	adapter := NewAdapter(id)

	// because we don't own the given stopper, we shouldn't call Add() on it
	subStopper := NewStopper()

	harmonyHubConnection := NewHarmonyHubConnection(addr, subStopper.Add())

	if err := harmonyHubConnection.InitAndAuthenticate(); err != nil {
		panic(err)
	}

	// does not actually go to that hostname/central service, but instead just the end device..
	// (bad name for stream recipient)
	if err := harmonyHubConnection.StartStreamTo("connect.logitech.com"); err != nil {
		panic(err)
	}

	if err := harmonyHubConnection.Bind(); err != nil {
		panic(err)
	}

	go func() {
		defer stopper.Done()

		log.Println("HarmonyHubAdapter: started")

		for {
			select {
			case <-stopper.ShouldStop:
				log.Println("HarmonyHubAdapter: stopping")
				harmonyHubConnection.EndStream()
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
