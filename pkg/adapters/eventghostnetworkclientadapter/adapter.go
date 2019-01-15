package eventghostnetworkclientadapter

import (
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/pkg/eventghostnetworkclient"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
)

func Start(adapter *hapitypes.Adapter, stop *stopper.Stopper) error {
	conn := eventghostnetworkclient.NewEventghostConnection(
		adapter.Conf.EventghostAddr,
		adapter.Conf.EventghostSecret)

	go func() {
		defer stop.Done()

		adapter.Logl.Info.Println("started")
		defer adapter.Logl.Info.Println("stopped")

		for {
			select {
			case <-stop.Signal:
				return
			case genericEvent := <-adapter.Outbound:
				switch e := genericEvent.(type) {
				case *hapitypes.PlaybackEvent:
					if err := conn.Send(e.Action, []string{}); err != nil {
						adapter.Logl.Error.Println(err.Error())
					}
				default:
					adapter.LogUnsupportedEvent(genericEvent)
				}
			}
		}
	}()

	return nil
}
