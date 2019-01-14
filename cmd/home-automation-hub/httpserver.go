package main

import (
	"encoding/json"
	"fmt"
	"github.com/function61/gokit/logger"
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
	"net/http"
)

func handleHttp(conf *hapitypes.ConfigFile, stop *stopper.Stopper) {
	log := logger.New("handleHttp")

	defer stop.Done()
	srv := &http.Server{Addr: ":8080"}

	go func() {
		<-stop.Signal

		log.Info("stopping HTTP")

		_ = srv.Shutdown(nil)
	}()

	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.Encode(conf)
	})

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Error(fmt.Sprintf("ListenAndServe(): %s", err.Error()))
	}
}
