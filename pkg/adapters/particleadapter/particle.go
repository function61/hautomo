package particleadapter

import (
	"context"
	"errors"

	"github.com/function61/hautomo/pkg/hapitypes"
	"github.com/function61/hautomo/pkg/particleapi"
)

func Start(ctx context.Context, adapter *hapitypes.Adapter) error {
	if adapter.Conf.ParticleAccessToken == "" || adapter.Conf.ParticleId == "" {
		return errors.New("ParticleAccessToken or ParticleId not defined")
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case genericEvent := <-adapter.Outbound:
			handleEvent(genericEvent, adapter)
		}
	}
}

func handleEvent(genericEvent hapitypes.OutboundEvent, adapter *hapitypes.Adapter) {
	switch e := genericEvent.(type) {
	case *hapitypes.PowerMsg:
		if err := particleapi.Invoke(adapter.Conf.ParticleId, "rf", e.PowerCommand, adapter.Conf.ParticleAccessToken); err != nil {
			adapter.Logl.Error.Println(err.Error())
		}
	default:
		adapter.LogUnsupportedEvent(genericEvent)
	}
}
