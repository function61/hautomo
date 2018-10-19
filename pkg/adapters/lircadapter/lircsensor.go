package lircadapter

import (
	"bufio"
	"fmt"
	"github.com/function61/gokit/logger"
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/hapitypes"
	"github.com/function61/home-automation-hub/pkg/signalfabric"
	"io"
	"os/exec"
	"regexp"
)

var log = logger.New("lircadapter")

// match lines like this: "000000037ff07bee 00 KEY_VOLUMEDOWN mceusb"

var mceUsbCommandRe = regexp.MustCompile(" 00 ([a-zA-Z_0-9]+) devinput$")

// reads LIRC's "$ irw" output
func StartSensor(fabric *signalfabric.Fabric, stop *stopper.Stopper) {
	defer stop.Done()

	log.Info("started")
	defer log.Info("stopped")

	irw := exec.Command("irw")

	stdoutPipe, pipeErr := irw.StdoutPipe()
	if pipeErr != nil {
		panic(pipeErr)
	}

	startErr := irw.Start()
	if startErr != nil {
		panic(startErr)
	}

	bufferedReader := bufio.NewReader(stdoutPipe)

	go func() {
		for {
			// TODO: implement isPrefix
			line, _, err := bufferedReader.ReadLine()
			if err != nil {
				if err != io.EOF {
					panic(err)
				}

				return // EOF
			}

			// "000000037ff07bee 00 KEY_VOLUMEDOWN mceusb" => "KEY_VOLUMEDOWN"
			mceUsbCommand := mceUsbCommandRe.FindStringSubmatch(string(line))
			if mceUsbCommand == nil {
				log.Error("mismatched command format")
				continue
			}

			ir := hapitypes.NewInfraredEvent("mceusb", mceUsbCommand[1])

			log.Debug(fmt.Sprintf("received %s", ir.Event))

			fabric.InfraredEvent <- ir
		}
	}()

	// TODO: do this via context cancel?
	go func() {
		<-stop.Signal

		log.Info("stopping")

		irw.Process.Kill()
	}()

	// wait to complete
	err := irw.Wait()

	log.Error(fmt.Sprintf("$ irw exited, error: %s", err.Error()))
}
