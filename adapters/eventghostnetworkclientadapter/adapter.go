package eventghostnetworkclientadapter

import (
	"github.com/function61/home-automation-hub/hapitypes"
	"github.com/function61/home-automation-hub/libraries/eventghostnetworkclient"
	"github.com/function61/home-automation-hub/util/stopper"
	"log"
)

func New(adapter *hapitypes.Adapter, config hapitypes.AdapterConfig, stopper *stopper.Stopper) *hapitypes.Adapter {
	conn := eventghostnetworkclient.NewEventghostConnection(
		config.EventghostAddr,
		config.EventghostSecret)

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
