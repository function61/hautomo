package dummyadapter

// dummy adapter just discards and logs messages

import (
	"github.com/function61/gokit/logger"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
)

var log = logger.New("dummyadapter")

func New(adapter *hapitypes.Adapter, config hapitypes.AdapterConfig) {
	go func() {
		log.Info("started")

		for {
			adapter.LogUnsupportedEvent(<-adapter.Event, log)
		}
	}()
}
