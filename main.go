package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"
)

const (
	rfParticleId = "310027000647343138333038"

	// FIXME: temporary
	sofaLight    = "d2ff0882"
	speakerLight = "98d3cb01"
)

var errDeviceNotFound = errors.New("device not found")

type Device struct {
	Id          string
	Name        string
	Description string
	// probably turned on if true
	// might be turned on even if false,
	ProbablyTurnedOn   bool
	ParticleOnCommand  string
	ParticleOffCommand string
}

func NewDevice(id string, name string, description string, particleOnCommand string, particleOffCommand string) *Device {
	return &Device{
		Id:                 id,
		Name:               name,
		Description:        description,
		ProbablyTurnedOn:   false,
		ParticleOnCommand:  particleOnCommand,
		ParticleOffCommand: particleOffCommand,
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

func (d *Device) TurnOn(app *Application) error {
	if d.Id == "c0730bb2" { // FIXME
		return app.harmonyHubConnection.HoldAndRelease("47917687", "PowerOn")
	}

	particleAccessToken, err := getParticleAccessToken()
	if err != nil {
		return err
	}

	return particleRequest(rfParticleId, "rf", d.ParticleOnCommand, particleAccessToken)
}

func (d *Device) TurnOff(app *Application) error {
	if d.Id == "c0730bb2" { // FIXME
		return app.harmonyHubConnection.HoldAndRelease("47917687", "PowerOff")
	}

	particleAccessToken, err := getParticleAccessToken()
	if err != nil {
		return err
	}

	return particleRequest(rfParticleId, "rf", d.ParticleOffCommand, particleAccessToken)
}

type Application struct {
	harmonyHubConnection *HarmonyHubConnection
	deviceById           map[string]*Device
	deviceGroupById      map[string]*DeviceGroup
}

func NewApplication() *Application {
	harmonyHubConnection := NewHarmonyHubConnection("192.168.1.153:5222")

	if err := harmonyHubConnection.InitAndAuthenticate(); err != nil {
		panic(err)
	}

	// does not actually go to that hostname/central service, but instead just the end device..
	// (bad name for stream recipient)
	if err := harmonyHubConnection.StartStreamTo("connect.logitech.com"); err != nil {
		panic(err)
	}

	if err := harmonyHubConnection.Bind(); err != nil {
		panic(err)
	}

	/* TODO: on close
	if err := harmonyHubConnection.EndStream(); err != nil {
		panic(err)
	}
	*/

	return &Application{
		harmonyHubConnection: harmonyHubConnection,
		deviceById:           make(map[string]*Device),
		deviceGroupById:      make(map[string]*DeviceGroup),
	}
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

	reqErr := device.TurnOn(a)

	if reqErr != nil {
		log.Printf("TurnOn error: %s", reqErr.Error())
	} else {
		device.ProbablyTurnedOn = true
	}

	return reqErr
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

	reqErr := device.TurnOff(a)

	if reqErr != nil {
		log.Printf("TurnOff error: %s", reqErr.Error())
	} else {
		device.ProbablyTurnedOn = false
	}

	return reqErr
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

func main() {
	var irw *bool = flag.Bool("irw", false, "infrared reading")
	var help *bool = flag.Bool("help", false, "help")

	flag.Parse()

	if *help {
		flag.PrintDefaults()
		return
	}

	app := NewApplication()

	// living room

	app.AttachDevice(NewDevice("c0730bb2", "Amplifier", "Onkyo TX-NR515", "", "")) // FIXME
	app.AttachDevice(NewDevice("d2ff0882", "Sofa light", "Floor light next the sofa", "C21", "C20"))
	app.AttachDevice(NewDevice("98d3cb01", "Speaker light", "Floor light under the speaker", "C31", "C30"))
	app.AttachDeviceGroup(NewDeviceGroup("cfb1b27f", "Living room lights", []string{
		"d2ff0882",
		"98d3cb01",
	}))

	app.SyncToCloud()

	if *irw {
		go irwPoller(app)
	}

	pollerLoop(app)
}
