package eventghostnetworkclientadapter

import (
	"github.com/function61/gokit/logger"
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/hapitypes"
	"github.com/function61/home-automation-hub/libraries/eventghostnetworkclient"
)

var log = logger.New("EventGhost")

func New(adapter *hapitypes.Adapter, config hapitypes.AdapterConfig, stop *stopper.Stopper) {
	conn := eventghostnetworkclient.NewEventghostConnection(
		config.EventghostAddr,
		config.EventghostSecret)

	go func() {
		defer stop.Done()

		log.Info("started")
		defer log.Info("stopped")

		for {
			select {
			case <-stop.Signal:
				return
			case playbackMsg := <-adapter.PlaybackMsg:
				if err := conn.Send(playbackMsg.Action, []string{}); err != nil {
					log.Error(err.Error())
				}
			}
		}
	}()
}
