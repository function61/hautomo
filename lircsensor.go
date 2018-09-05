package main

import (
	"bufio"
	"github.com/function61/home-automation-hub/hapitypes"
	"github.com/function61/home-automation-hub/util/stopper"
	"io"
	"log"
	"os/exec"
	"regexp"
)

// match lines like this: "000000037ff07bee 00 KEY_VOLUMEDOWN mceusb"

var mceUsbCommandRe = regexp.MustCompile(" 00 ([a-zA-Z_0-9]+) devinput$")

// reads LIRC's "$ irw" output
func irwPoller(app *Application, stopper *stopper.Stopper) {
	defer stopper.Done()

	log.Println("irwPoller: starting")

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
				log.Println("irwPoller: mismatched command format")
				continue
			}

			ir := hapitypes.NewInfraredEvent("mceusb", mceUsbCommand[1])

			log.Printf("irwPoller: received %s", ir.Event)

			app.infraredEvent <- ir
		}
	}()

	go func() {
		<-stopper.ShouldStop

		log.Println("irwPoller: asked to stop")

		irw.Process.Kill()
	}()

	// wait to complete
	err := irw.Wait()

	log.Printf("irwPoller: child exited, error: %s", err.Error())
}
