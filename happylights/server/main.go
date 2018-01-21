package main

import (
	"../../util/systemdinstaller"
	"../types"
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
)

// compile this with:
// $ GOOS=linux GOARCH=arm go build -o server_arm

// controls happylights over Bluetooth BLE
func buildHappylightBluetoothRequestCmd(req types.LightRequest) *exec.Cmd {
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

	return exec.Command(
		"gatttool",
		"-i", "hci0",
		"-b", req.BluetoothAddr,
		"--char-write-req",
		"-a", "0x0021", // known handle for light request
		"-n", reqHex)
}

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--help" {
		fmt.Printf("Usage: %s [--write-systemd-unit-file]\n", os.Args[0])
		return
	}
	if len(os.Args) == 2 && os.Args[1] == "--write-systemd-unit-file" {
		if err := systemdinstaller.InstallSystemdServiceFile("happylights", "happylights server daemon"); err != nil {
			panic(err)
		}
		return
	}
	if len(os.Args) != 1 {
		fmt.Printf("Unknown args, please run %s --help\n", os.Args[0])
		os.Exit(1)
	}

	// listen to incoming udp packets
	pc, err := net.ListenPacket("udp", "0.0.0.0:9092")
	if err != nil {
		log.Fatal(err)
	}
	defer pc.Close()

	for {
		bytesRaw := make([]byte, 4096)

		_, _, err := pc.ReadFrom(bytesRaw)
		if err != nil {
			panic(err)
		}

		dec := gob.NewDecoder(bytes.NewBuffer(bytesRaw))
		var lightRequest types.LightRequest
		if err := dec.Decode(&lightRequest); err != nil {
			panic(err)
		}

		happylightCmd := buildHappylightBluetoothRequestCmd(lightRequest)

		output, err := happylightCmd.CombinedOutput()
		if err != nil {
			log.Printf("Error %s, stdout: %s", err.Error(), output)
		}
	}
}
