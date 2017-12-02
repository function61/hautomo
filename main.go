package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/function61/eventhorizon/util/clicommon"
	"log"
	"strings"
)

const (
	// FIXME: temporary
	sofaLight    = "d2ff0882"
	speakerLight = "98d3cb01"
)

var errDeviceNotFound = errors.New("device not found")

type Device struct {
	Id          string
	Name        string
	Description string

	// adapter details
	AdapterId        string
	AdaptersDeviceId string // id by which the adapter identifies this device

	// probably turned on if true
	// might be turned on even if false,
	ProbablyTurnedOn bool

	PowerOnCmd  string
	PowerOffCmd string
}

func NewDevice(id string, adapterId string, adaptersDeviceId string, name string, description string, powerOnCmd string, powerOffCmd string) *Device {
	return &Device{
		Id:          id,
		Name:        name,
		Description: description,

		AdapterId:        adapterId,
		AdaptersDeviceId: adaptersDeviceId,

		// state
		ProbablyTurnedOn: false,

		PowerOnCmd:  powerOnCmd,
		PowerOffCmd: powerOffCmd,
	}
}

type DeviceGroup struct {
	Id        string
	Name      string
	DeviceIds []string
}

func NewDeviceGroup(id string, name string, deviceIds []string) *DeviceGroup {
	return &DeviceGroup{
		Id:        id,
		Name:      name,
		DeviceIds: deviceIds,
	}
}

type Application struct {
	adapterById     map[string]*Adapter
	deviceById      map[string]*Device
	deviceGroupById map[string]*DeviceGroup
}

func NewApplication() *Application {
	return &Application{
		adapterById:     make(map[string]*Adapter),
		deviceById:      make(map[string]*Device),
		deviceGroupById: make(map[string]*DeviceGroup),
	}
}

func (a *Application) DefineAdapter(adapter *Adapter) {
	a.adapterById[adapter.Id] = adapter
}

func (a *Application) AttachDevice(device *Device) {
	a.deviceById[device.Id] = device
}

func (a *Application) AttachDeviceGroup(deviceGroup *DeviceGroup) {
	a.deviceGroupById[deviceGroup.Id] = deviceGroup
}

func (a *Application) TurnOn(deviceId string) error {
	device, deviceFound := a.deviceById[deviceId]
	if !deviceFound {
		deviceGroup, deviceGroupFound := a.deviceGroupById[deviceId]
		if !deviceGroupFound {
			return errDeviceNotFound
		}

		return a.turnOnDeviceGroup(deviceGroup)
	}

	log.Printf("TurnOn: %s", device.Name)

	adapter := a.adapterById[device.AdapterId]
	adapter.PowerMsg <- NewPowerMsg(device.AdaptersDeviceId, device.PowerOnCmd)

	device.ProbablyTurnedOn = true

	return nil
}

func (a *Application) TurnOff(deviceId string) error {
	device, deviceFound := a.deviceById[deviceId]
	if !deviceFound {
		deviceGroup, deviceGroupFound := a.deviceGroupById[deviceId]
		if !deviceGroupFound {
			return errDeviceNotFound
		}

		return a.turnOffDeviceGroup(deviceGroup)
	}

	log.Printf("TurnOff: %s", device.Name)

	adapter := a.adapterById[device.AdapterId]
	adapter.PowerMsg <- NewPowerMsg(device.AdaptersDeviceId, device.PowerOffCmd)

	device.ProbablyTurnedOn = false

	return nil
}

func (a *Application) turnOnDeviceGroup(deviceGroup *DeviceGroup) error {
	log.Printf("turnOnDeviceGroup: %s", deviceGroup.Name)

	for _, deviceId := range deviceGroup.DeviceIds {
		device := a.deviceById[deviceId] // TODO: panics if not found

		if device.ProbablyTurnedOn {
			continue
		}

		_ = a.TurnOn(device.Id)
	}

	return nil
}

func (a *Application) turnOffDeviceGroup(deviceGroup *DeviceGroup) error {
	log.Printf("turnOffDeviceGroup: %s", deviceGroup.Name)

	for _, deviceId := range deviceGroup.DeviceIds {
		device := a.deviceById[deviceId] // TODO: panics if not found

		// intentionally missing ProbablyTurnedOn check

		_ = a.TurnOff(device.Id)
	}

	return nil
}

func (a *Application) SyncToCloud() {
	lines := []string{""} // empty line to start output from next log line

	for _, device := range a.deviceById {
		lines = append(lines, fmt.Sprintf("createDevice('%s', '%s', '%s'),",
			device.Id,
			device.Name,
			device.Description))
	}

	for _, deviceGroup := range a.deviceGroupById {
		lines = append(lines, fmt.Sprintf("createDevice('%s', '%s', '%s'),",
			deviceGroup.Id,
			deviceGroup.Name,
			"Device group: "+deviceGroup.Name))
	}

	log.Println(strings.Join(lines, "\n"))
}

type PowerMsg struct {
	DeviceId     string
	PowerCommand string
}

func NewPowerMsg(deviceId string, powerCommand string) PowerMsg {
	return PowerMsg{
		DeviceId:     deviceId,
		PowerCommand: powerCommand,
	}
}

type Adapter struct {
	Id       string
	PowerMsg chan PowerMsg
}

func NewAdapter(id string) *Adapter {
	return &Adapter{
		Id:       id,
		PowerMsg: make(chan PowerMsg),
	}
}

func main() {
	var irw *bool = flag.Bool("irw", false, "infrared reading")

	flag.Parse()

	stopper := NewStopper()
	app := NewApplication()

	app.DefineAdapter(NewHarmonyHubAdapter("harmonyHubAdapter", "192.168.1.153:5222", stopper.Add()))

	particleAccessToken := getParticleAccessToken()

	app.DefineAdapter(NewParticleAdapter("particleAdapter", "310027000647343138333038", particleAccessToken))

	app.AttachDevice(NewDevice("c0730bb2", "harmonyHubAdapter", "47917687", "Amplifier", "Onkyo TX-NR515", "PowerOn", "PowerOff"))

	// for some reason the TV only wakes up with PowerToggle, not PowerOn
	app.AttachDevice(NewDevice("7e7453da", "harmonyHubAdapter", "????", "TV", `Philips 55" 4K 55PUS7909`, "PowerToggle", "PowerOff"))

	app.AttachDevice(NewDevice("d2ff0882", "particleAdapter", "", "Sofa light", "Floor light next the sofa", "C21", "C20"))
	app.AttachDevice(NewDevice("98d3cb01", "particleAdapter", "", "Speaker light", "Floor light under the speaker", "C31", "C30"))

	app.AttachDeviceGroup(NewDeviceGroup("cfb1b27f", "Living room lights", []string{
		"d2ff0882",
		"98d3cb01",
	}))

	app.SyncToCloud()

	if *irw {
		go irwPoller(app, stopper.Add())
	}

	go sqsPollerLoop(app, stopper.Add())

	clicommon.WaitForInterrupt()

	log.Println("main: received interrupt")

	stopper.StopAll()

	log.Println("main: all components stopped")
}
