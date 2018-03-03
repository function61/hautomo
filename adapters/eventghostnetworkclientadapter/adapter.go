package eventghostnetworkclientadapter

import (
	"../../hapitypes"
	"../../libraries/eventghostnetworkclient"
	"../../util/stopper"
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
