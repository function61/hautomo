package dummyadapter

// dummy adapter just discards and logs messages

import (
	"context"

	"github.com/function61/hautomo/pkg/hapitypes"
)

func Start(ctx context.Context, adapter *hapitypes.Adapter) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case e := <-adapter.Outbound:
			adapter.LogUnsupportedEvent(e)
		}
	}
}
