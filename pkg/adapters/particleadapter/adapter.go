package particleadapter

import (
	"github.com/function61/gokit/logger"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
	"github.com/function61/home-automation-hub/pkg/particleapi"
)

var log = logger.New("particleadapter")

func New(adapter *hapitypes.Adapter, config hapitypes.AdapterConfig) {
	go func() {
		log.Info("started")

		for {
			select {
			case powerMsg := <-adapter.PowerMsg:
				if config.ParticleAccessToken == "" || config.ParticleId == "" {
					log.Error("ParticleAccessToken or ParticleId not defined")
					continue
				}

				if err := particleapi.Invoke(config.ParticleId, "rf", powerMsg.PowerCommand, config.ParticleAccessToken); err != nil {
					log.Error(err.Error())
				}
			}
		}
	}()
}
