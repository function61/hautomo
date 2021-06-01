package ezstack

// captures ZNP frames and writes them to a textual log file so you can reverse-engineer
// communications / capture messages for use with tests.
//
// TODO: research if we can utilize pcap-ng format so the packets could be opened in Wireshark?

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/function61/hautomo/pkg/ezstack/znp"
)

func runPacketCapture(ctx context.Context, packetCaptureFilename string, networkProcessor *znp.Znp) error {
	packetCaptureFile, err := os.Create(packetCaptureFilename)
	if err != nil {
		return err
	}
	defer packetCaptureFile.Close()

	for {
		select {
		case <-ctx.Done():
			return packetCaptureFile.Close() // double close intentional
		case frame := <-networkProcessor.InFramesLog():
			if _, err := fmt.Fprintf(
				packetCaptureFile,
				"%s CommandType=%d Subsystem=%d Command=%d Payload=%x\n",
				time.Now().Format(time.RFC3339Nano),
				frame.CommandType,
				frame.Subsystem,
				frame.Command,
				frame.Payload,
			); err != nil {
				return err
			}
		}
	}
}
