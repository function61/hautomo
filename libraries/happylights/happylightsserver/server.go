package happylightsserver

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/function61/gokit/retry"
	"github.com/function61/home-automation-hub/libraries/happylights/types"
	"log"
	"net"
	"os/exec"
	"time"
)

// controls happylights over Bluetooth BLE
func buildHappylightBluetoothRequestCmd(ctx context.Context, req types.LightRequest) *exec.Cmd {
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

	// do not let one attempt last more than this
	ctxCmd, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return exec.CommandContext(
		ctxCmd,
		"gatttool",
		"-i", "hci0",
		"-b", req.BluetoothAddr,
		"--char-write-req",
		"-a", "0x0021", // known handle for light request
		"-n", reqHex)
}

func runServer() {
	listenAddr := "0.0.0.0:9092"

	log.Printf("Starting to listen on %s", listenAddr)

	happylightIncomingCmdSock, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer happylightIncomingCmdSock.Close()

	for {
		bytesRaw := make([]byte, 4096)

		_, _, err := happylightIncomingCmdSock.ReadFrom(bytesRaw)
		if err != nil {
			log.Printf("runServer: ReadFrom() failed: %s", err.Error())
			time.Sleep(500 * time.Millisecond) // prevent hot loop
			continue
		}

		gobDecoder := gob.NewDecoder(bytes.NewBuffer(bytesRaw))
		lightRequest := types.LightRequest{}
		if err := gobDecoder.Decode(&lightRequest); err != nil {
			log.Printf("runServer: GOB Decode() failed: %s", err.Error())
			continue
		}

		ifFails := func(err error) {
			log.Printf("happylightCmd %s", err.Error())
		}

		// try for 15 seconds
		ctx, cancel := context.WithTimeout(context.TODO(), 15*time.Second)
		defer cancel()

		retry.Retry(ctx, func(ctx context.Context) error {
			happylightCmd := buildHappylightBluetoothRequestCmd(ctx, lightRequest)

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
}
