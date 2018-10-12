package main

import (
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/hapitypes"
	"log"
	"time"
)

func infraredSimulator(app *Application, key string, stop *stopper.Stopper) {
	defer stop.Done()

	log.Println("IR simulator: started")

	for {
		select {
		case <-time.After(5 * time.Second):
			app.infraredEvent <- hapitypes.NewInfraredEvent("simulated_remote", key)
		case <-stop.Signal:
			log.Println("IR simulator: stopping")
			return
		}
	}
}
