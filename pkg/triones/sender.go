package triones

// controls Triones (sometimes marketed as Happylights) over Bluetooth BLE

import (
	"context"
	"fmt"
	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/retry"
	"log"
	"os/exec"
	"time"
)

func Send(ctx context.Context, req Request, logger *log.Logger) error {
	ifFails := func(err error) {
		// retry has enough context about "attempt failure"
		logex.Levels(logger).Error.Println(err.Error())
	}

	gatttool := gattToolArgs(req.BluetoothAddr, requestToHex(req))

	return retry.Retry(ctx, func(ctx context.Context) error {
		// do not let one attempt last more than this
		ctxCmd, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		gatttoolCmd := exec.CommandContext(
			ctxCmd,
			gatttool[0],
			gatttool[1:]...)

		output, err := gatttoolCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf( // retryer adds sufficient log prefix
				"%s, stdout: %s",
				err.Error(),
				output)
		}

		return nil
	}, retry.DefaultBackoff(), ifFails)
}

// if running into errors:
// https://stackoverflow.com/questions/22062037/hcitool-lescan-shows-i-o-error
func gattToolArgs(btAddr string, reqHex string) []string {
	return []string{
		"gatttool",
		"-i", "hci0",
		"-b", btAddr,
		"--char-write-req",
		"-a", "0x0021", // known handle for light request
		"-n", reqHex,
	}
}

// thanks https://github.com/madhead/saberlight/blob/master/protocols/Triones/protocol.md
// for the AWESOME protocol description! :)
func requestToHex(req Request) string {
	switch req.Kind {
	case RequestKindOn:
		return "cc2333"
	case RequestKindOff:
		return "cc2433"
	case RequestKindRGB:
		return fmt.Sprintf(
			"56%02x%02x%02x00f0aa",
			req.RgbOpts.Red,
			req.RgbOpts.Green,
			req.RgbOpts.Blue)
	case RequestKindWhite:
		return fmt.Sprintf(
			"56000000%02x0faa",
			req.WhiteOpts.Brightness)
	default:
		panic("unkown request kind")
	}
}
