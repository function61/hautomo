package devicegroupadapter

import (
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
)

// this adapter just basically copies the outbound event as multiple copies with rewritten
// device ID and posts it as inbound again

func Start(adapter *hapitypes.Adapter, stop *stopper.Stopper) error {
	go func() {
		defer stop.Done()
		adapter.Logl.Info.Println("started")
		defer adapter.Logl.Info.Println("stopped")

		for {
			select {
			case <-stop.Signal:
				return
			case event := <-adapter.Outbound:
				for _, to := range adapter.Conf.DevicegroupDevices {
					adapter.Receive(event.RedirectInbound(to))
				}
			}
		}
	}()

	return nil
}
