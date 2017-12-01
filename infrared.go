package main

import (
	"bufio"
	"io"
	"log"
	"os/exec"
	"regexp"
)

// match lines like this: "000000037ff07bee 00 KEY_VOLUMEDOWN mceusb"

var mceUsbCommandRe = regexp.MustCompile(" 00 ([a-zA-Z_0-9]+) mceusb$")

// reads LIRC's "$ irw" output
func irwPoller(app *Application, stopper *Stopper) {
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

			log.Printf("irwPoller: received %s", mceUsbCommand[1])

			switch mceUsbCommand[1] {
			case "KEY_VOLUMEUP":
				app.TurnOn(speakerLight)
			case "KEY_VOLUMEDOWN":
				app.TurnOff(speakerLight)
			case "KEY_CHANNELUP":
				app.TurnOn(sofaLight)
			case "KEY_CHANNELDOWN":
				app.TurnOff(sofaLight)
			default:
				log.Println("irwPoller: command ignored")
			}

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
