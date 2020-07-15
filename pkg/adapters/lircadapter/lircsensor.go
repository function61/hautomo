package lircadapter

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"regexp"

	"github.com/function61/hautomo/pkg/hapitypes"
)

// match lines like this: "000000037ff07bee 00 KEY_VOLUMEDOWN mceusb"

var irwOutputParseRe = regexp.MustCompile(`^[0-9a-f]{16} 00 ([^ ]+) (.+)$`)

func irwOutputLineToIrEvent(line string) *hapitypes.RawInfraredEvent {
	irCommand := irwOutputParseRe.FindStringSubmatch(line)
	if irCommand == nil {
		return nil
	}

	return hapitypes.NewRawInfraredEvent(irCommand[2], irCommand[1])
}

// reads LIRC's "$ irw" output
func Start(ctx context.Context, adapter *hapitypes.Adapter) error {
	irw := exec.CommandContext(ctx, "irw")

	stdoutPipe, err := irw.StdoutPipe()
	if err != nil {
		return err
	}

	if err := irw.Start(); err != nil {
		return err
	}

	go func() {
		bufferedReader := bufio.NewReader(stdoutPipe)

		for {
			// TODO: implement isPrefix
			line, _, err := bufferedReader.ReadLine()
			if err != nil {
				if err == io.EOF {
					return
				}

				panic(err)
			}

			// "000000037ff07bee 00 KEY_VOLUMEDOWN mceusb" => "KEY_VOLUMEDOWN"
			irEvent := irwOutputLineToIrEvent(string(line))
			if irEvent == nil {
				adapter.Logl.Error.Printf("mismatched command format: %s", string(line))
				continue
			}

			adapter.Receive(irEvent)
		}
	}()

	// should stop on context cancellation the latest
	err = irw.Wait()

	select {
	case <-ctx.Done():
		return nil // stopping was expected
	default:
		return fmt.Errorf("irw exited: %w", err)
	}
}
