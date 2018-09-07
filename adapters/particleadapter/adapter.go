package particleadapter

import (
	"github.com/function61/home-automation-hub/hapitypes"
	"github.com/function61/home-automation-hub/libraries/particleapi"
	"log"
)

func New(id string, particleId string, accessToken string) *hapitypes.Adapter {
	adapter := hapitypes.NewAdapter(id)

	go func() {
		log.Println("particleadapter: started")

		for {
			select {
			case powerMsg := <-adapter.PowerMsg:
				log.Printf("particleadapter: got PowerMsg")

				if accessToken == "" {
					log.Printf("particleadapter: error: accessToken not defined")
					continue
				}
				if err := particleapi.Invoke(particleId, "rf", powerMsg.PowerCommand, accessToken); err != nil {
					log.Printf("particleadapter: request failed: %s", err.Error())
				}
			}
		}
	}()

	return adapter
}
