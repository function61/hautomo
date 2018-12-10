package eventghostnetworkclientadapter

import (
	"github.com/function61/gokit/logger"
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/pkg/eventghostnetworkclient"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
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
			case genericEvent := <-adapter.Event:
				switch e := genericEvent.(type) {
				case *hapitypes.PlaybackEvent:
					if err := conn.Send(e.Action, []string{}); err != nil {
						log.Error(err.Error())
					}
				default:
					adapter.LogUnsupportedEvent(genericEvent, log)
				}
			}
		}
	}()
}
