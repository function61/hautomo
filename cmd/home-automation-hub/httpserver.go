package main

import (
	"context"
	"encoding/json"
	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
	"log"
	"net/http"
)

func handleHttp(conf *hapitypes.ConfigFile, logger *log.Logger, stop *stopper.Stopper) {
	logl := logex.Levels(logger)

	defer stop.Done()
	srv := &http.Server{Addr: ":8097"}

	go func() {
		<-stop.Signal

		logl.Info.Println("stopping HTTP")

		_ = srv.Shutdown(context.TODO())
	}()

	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.Encode(conf)
	})

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		logl.Error.Printf("ListenAndServe(): %s", err.Error())
	}
}
