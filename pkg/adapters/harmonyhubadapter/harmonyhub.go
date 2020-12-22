package harmonyhubadapter

import (
	"context"

	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/taskrunner"
	"github.com/function61/hautomo/pkg/hapitypes"
	"github.com/function61/hautomo/pkg/harmonyhub"
)

func Start(ctx context.Context, adapter *hapitypes.Adapter) error {
	harmonyhubEnableLogs := false

	harmonyhubLogger := logex.Prefix("lib", adapter.Log)
	if !harmonyhubEnableLogs {
		harmonyhubLogger = logex.Discard
	}

	harmonyHubConnection := harmonyhub.NewHarmonyHubConnection(
		ctx,
		adapter.Conf.Url,
		harmonyhubLogger)

	connTask := taskrunner.New(ctx, adapter.Log)
	connTask.Start("conn", func(ctx context.Context) error {
		return harmonyHubConnection.Task(ctx)
	})

	for {
		select {
		case <-ctx.Done():
			return connTask.Wait()
		case err := <-connTask.Done(): // subtask crash
			return err
		case genericEvent := <-adapter.Outbound:
			switch e := genericEvent.(type) {
			case *hapitypes.PowerMsg:
				if err := harmonyHubConnection.HoldAndRelease(e.DeviceId, e.PowerCommand); err != nil {
					adapter.Logl.Error.Printf("HoldAndRelease: %s", err.Error())
				}
			case *hapitypes.InfraredEvent:
				if err := harmonyHubConnection.HoldAndRelease(e.Device, e.Command); err != nil {
					adapter.Logl.Error.Printf("HoldAndRelease: %s", err.Error())
				}
			default:
				adapter.LogUnsupportedEvent(genericEvent)
			}
		}
	}
}
