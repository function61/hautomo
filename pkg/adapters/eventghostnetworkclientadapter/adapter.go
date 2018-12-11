package eventghostnetworkclientadapter

import (
	"github.com/function61/gokit/logger"
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/pkg/eventghostnetworkclient"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
)

var log = logger.New("EventGhost")

func Start(adapter *hapitypes.Adapter, stop *stopper.Stopper) error {
	conn := eventghostnetworkclient.NewEventghostConnection(
		adapter.Conf.EventghostAddr,
		adapter.Conf.EventghostSecret)

	go func() {
		defer stop.Done()

		log.Info("started")
		defer log.Info("stopped")

		for {
			select {
			case <-stop.Signal:
				return
			case genericEvent := <-adapter.Outbound:
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

	return nil
}
