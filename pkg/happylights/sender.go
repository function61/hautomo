package happylights

import (
	"context"
	"fmt"
	"github.com/function61/gokit/retry"
	"log"
	"os/exec"
	"time"
)

// controls happylights over Bluetooth BLE
func lightRequestToGattToolArgs(req LightRequest) []string {
	reqHex := ""

	if req.IsOff() {
		// turns off
		reqHex = "cc2433"
	} else if req.IsOn() {
		// turns on
		reqHex = "cc2333"
	} else {
		// sets rgb. sadly does not turn on if currently off
		reqHex = fmt.Sprintf("56%02x%02x%02x00f0aa", req.Red, req.Green, req.Blue)
	}

	// if running into errors:
	// https://stackoverflow.com/questions/22062037/hcitool-lescan-shows-i-o-error

	return []string{
		"gatttool",
		"-i", "hci0",
		"-b", req.BluetoothAddr,
		"--char-write-req",
		"-a", "0x0021", // known handle for light request
		"-n", reqHex,
	}
}

func Send(ctx context.Context, req LightRequest) error {
	ifFails := func(err error) {
		log.Printf("happylightCmd %s", err.Error())
	}

	return retry.Retry(ctx, func(ctx context.Context) error {
		happylightCmdArgs := lightRequestToGattToolArgs(req)

		// do not let one attempt last more than this
		ctxCmd, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		happylightCmd := exec.CommandContext(
			ctxCmd,
			happylightCmdArgs[0],
			happylightCmdArgs[1:]...)

		output, err := happylightCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf( // retryer adds sufficient log prefix
				"%s, stdout: %s",
				err.Error(),
				output)
		}

		return nil
	}, retry.DefaultBackoff(), ifFails)
}
