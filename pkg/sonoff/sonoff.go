// controls Sonoff Basic relays using the Tasmota open source firmware
package sonoff

import (
	"context"
	"fmt"

	"github.com/function61/gokit/ezhttp"
)

func TurnOn(ctx context.Context, deviceAddr string) error {
	return onOrOff(ctx, "http://"+deviceAddr+"/cm?cmnd=Power%20On", "ON")
}

func TurnOff(ctx context.Context, deviceAddr string) error {
	return onOrOff(ctx, "http://"+deviceAddr+"/cm?cmnd=Power%20off", "OFF")
}

// https://github.com/arendst/Sonoff-Tasmota/wiki/Commands#sending-commands-with-web-requests
func onOrOff(ctx context.Context, endpoint string, expectedPowerState string) error {
	resp := struct {
		Power string `json:"POWER"` // ON | OFF
	}{}

	if _, err := ezhttp.Get(
		ctx,
		endpoint,
		ezhttp.RespondsJson(&resp, true)); err != nil {
		return err
	}

	if resp.Power != expectedPowerState {
		return fmt.Errorf("expecting POWER=%s got=%s", expectedPowerState, resp.Power)
	}

	return nil
}
