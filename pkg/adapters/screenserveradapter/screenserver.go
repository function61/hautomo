// Sends notifications to screens of screen-server
package screenserveradapter

import (
	"context"
	"errors"

	"github.com/function61/hautomo/pkg/hapitypes"
	"github.com/function61/hautomo/pkg/screenserverclient"
)

func Start(ctx context.Context, adapter *hapitypes.Adapter) error {
	if adapter.Conf.Url == "" {
		return errors.New("Url empty")
	}

	server := screenserverclient.Server(adapter.Conf.Url)

	for {
		select {
		case <-ctx.Done():
			return nil
		case genericEvent := <-adapter.Outbound:
			switch e := genericEvent.(type) {
			case *hapitypes.PowerMsg:
				screen := screenserverclient.ClientClient(e.DeviceId)

				if err := func() error {
					if e.On {
						return screen.ScreenPowerOn(context.TODO())
					} else {
						return screen.ScreenPowerOff(context.TODO())
					}
				}(); err != nil {
					adapter.Logl.Error.Println(err)
				}
			case *hapitypes.PlaySoundEvent:
				screen := screenserverclient.ClientClient(e.Device)

				if err := screen.PlayAudio(context.TODO(), e.Url); err != nil {
					adapter.Logl.Error.Printf("PlayAudio: %v", err)
				}
			case *hapitypes.NotificationEvent:
				// TODO: adapters device id?
				screen := server.Screen(e.Device)

				if err := screen.DisplayNotification(ctx, e.Message); err != nil {
					adapter.Logl.Error.Printf("DisplayNotification: %v", err)
				}
			default:
				adapter.LogUnsupportedEvent(genericEvent)
			}
		}
	}
}
