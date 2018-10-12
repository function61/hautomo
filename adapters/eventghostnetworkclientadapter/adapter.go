package eventghostnetworkclientadapter

import (
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/hapitypes"
	"github.com/function61/home-automation-hub/libraries/eventghostnetworkclient"
	"log"
)

func New(adapter *hapitypes.Adapter, config hapitypes.AdapterConfig, stop *stopper.Stopper) *hapitypes.Adapter {
	conn := eventghostnetworkclient.NewEventghostConnection(
		config.EventghostAddr,
		config.EventghostSecret)

	go func() {
		defer stop.Done()

		log.Println("eventghostnetworkclientadapter: started")

		for {
			select {
			case <-stop.Signal:
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
