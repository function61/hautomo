package particleadapter

import (
	"github.com/function61/home-automation-hub/hapitypes"
	"github.com/function61/home-automation-hub/libraries/particleapi"
	"log"
)

func New(adapter *hapitypes.Adapter, config hapitypes.AdapterConfig) *hapitypes.Adapter {
	go func() {
		log.Println("particleadapter: started")

		for {
			select {
			case powerMsg := <-adapter.PowerMsg:
				if config.ParticleAccessToken == "" || config.ParticleId == "" {
					log.Printf("particleadapter: error: ParticleAccessToken or ParticleId not defined")
					continue
				}

				if err := particleapi.Invoke(config.ParticleId, "rf", powerMsg.PowerCommand, config.ParticleId); err != nil {
					log.Printf("particleadapter: request failed: %s", err.Error())
				}
			}
		}
	}()

	return adapter
}
