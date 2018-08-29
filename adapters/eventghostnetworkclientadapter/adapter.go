package eventghostnetworkclientadapter

import (
	"github.com/function61/home-automation-hub/hapitypes"
	"github.com/function61/home-automation-hub/libraries/eventghostnetworkclient"
	"github.com/function61/home-automation-hub/util/stopper"
	"log"
)

func New(id string, addr string, secret string, stopper *stopper.Stopper) *hapitypes.Adapter {
	adapter := hapitypes.NewAdapter(id)

	conn := eventghostnetworkclient.NewEventghostConnection(addr, secret)

	go func() {
		defer stopper.Done()

		log.Println("eventghostnetworkclientadapter: started")

		for {
			select {
			case <-stopper.ShouldStop:
				log.Println("eventghostnetworkclientadapter: stopping")
				return
			case playbackMsg := <-adapter.PlaybackMsg:
				if err := conn.Send(playbackMsg.Action, []string{}); err != nil {
					log.Printf("eventghostnetworkclientadapter: error %s", err.Error())
				}
			}
		}
	}()

	return adapter
}
