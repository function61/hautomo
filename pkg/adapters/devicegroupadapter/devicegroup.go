package devicegroupadapter

import (
	"context"

	"github.com/function61/hautomo/pkg/hapitypes"
)

// this adapter just basically copies the outbound event as multiple copies with rewritten
// device ID and posts it as inbound again

func Start(ctx context.Context, adapter *hapitypes.Adapter) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case event := <-adapter.Outbound:
			for _, to := range adapter.Conf.DevicegroupDevices {
				adapter.Receive(event.RedirectInbound(to))
			}
		}
	}
}
