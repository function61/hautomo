package main

import (
	"github.com/function61/home-automation-hub/hapitypes"
	"github.com/function61/home-automation-hub/util/stopper"
	"log"
	"time"
)

func infraredSimulator(app *Application, key string, stopper *stopper.Stopper) {
	defer stopper.Done()

	log.Println("IR simulator: started")

	for {
		select {
		case <-time.After(5 * time.Second):
			app.infraredEvent <- hapitypes.NewInfraredEvent("simulated_remote", key)
		case <-stopper.ShouldStop:
			log.Println("IR simulator: stopping")
			return
		}
	}
}
