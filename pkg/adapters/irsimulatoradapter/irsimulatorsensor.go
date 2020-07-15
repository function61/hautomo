package irsimulatoradapter

import (
	"context"
	"time"

	"github.com/function61/hautomo/pkg/hapitypes"
)

func Start(ctx context.Context, adapter *hapitypes.Adapter) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(5 * time.Second):
			adapter.Receive(hapitypes.NewRawInfraredEvent(
				"simulated_remote",
				adapter.Conf.IrSimulatorKey))
		}
	}
}
