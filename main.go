package main

import (
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

type Application struct {
	adapterById     map[string]*Adapter
	deviceById      map[string]*Device
	deviceGroupById map[string]*DeviceGroup
	infraredEvent   chan InfraredEvent
}

func NewApplication(stopper *Stopper) *Application {
	app := &Application{
		adapterById:     make(map[string]*Adapter),
		deviceById:      make(map[string]*Device),
		deviceGroupById: make(map[string]*DeviceGroup),
		infraredEvent:   make(chan InfraredEvent),
	}

	go func() {
		defer stopper.Done()

		log.Println("application: started")

		for {
			select {
			case <-stopper.ShouldStop:
				log.Println("application: stopping")
				return
			case ir := <-app.infraredEvent:
				log.Printf("application: IR: %s", ir.Event)

				switch ir.Event {
				case "KEY_VOLUMEUP":
					app.TurnOn(app.deviceById[speakerLight])
				case "KEY_VOLUMEDOWN":
					app.TurnOff(app.deviceById[speakerLight])
				case "KEY_CHANNELUP":
					app.TurnOn(app.deviceById[sofaLight])
				case "KEY_CHANNELDOWN":
					app.TurnOff(app.deviceById[sofaLight])
				default:
					log.Println("application: IR ignored")
				}
			}
		}

	}()

	return app
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

func (a *Application) TurnOnDeviceOrDeviceGroup(deviceId string) error {
	device, deviceFound := a.deviceById[deviceId]
	if deviceFound {
		return a.TurnOn(device)

	}

	deviceGroup, deviceGroupFound := a.deviceGroupById[deviceId]
	if deviceGroupFound {
		return a.turnOnDeviceGroup(deviceGroup)
	}

	return errDeviceNotFound
}

func (a *Application) TurnOffDeviceOrDeviceGroup(deviceId string) error {
	device, deviceFound := a.deviceById[deviceId]
	if deviceFound {
		return a.TurnOff(device)

	}

	deviceGroup, deviceGroupFound := a.deviceGroupById[deviceId]
	if deviceGroupFound {
		return a.turnOffDeviceGroup(deviceGroup)
	}

	return errDeviceNotFound
}

func (a *Application) TurnOn(device *Device) error {
	log.Printf("TurnOn: %s", device.Name)

	adapter := a.adapterById[device.AdapterId]
	adapter.PowerMsg <- NewPowerMsg(device.AdaptersDeviceId, device.PowerOnCmd)

	device.ProbablyTurnedOn = true

	return nil
}

func (a *Application) TurnOff(device *Device) error {
	log.Printf("TurnOff: %s", device.Name)

	adapter := a.adapterById[device.AdapterId]
	adapter.PowerMsg <- NewPowerMsg(device.AdaptersDeviceId, device.PowerOffCmd)

	device.ProbablyTurnedOn = false

	return nil
}

func (a *Application) turnOnDeviceGroup(deviceGroup *DeviceGroup) error {
	log.Printf("turnOnDeviceGroup: %s", deviceGroup.Name)

	for _, deviceId := range deviceGroup.DeviceIds {
		device := a.deviceById[deviceId] // FIXME: panics if not found

		if device.ProbablyTurnedOn {
			continue
		}

		_ = a.TurnOn(device)
	}

	return nil
}

func (a *Application) turnOffDeviceGroup(deviceGroup *DeviceGroup) error {
	log.Printf("turnOffDeviceGroup: %s", deviceGroup.Name)

	for _, deviceId := range deviceGroup.DeviceIds {
		device := a.deviceById[deviceId] // FIXME: panics if not found

		// intentionally missing ProbablyTurnedOn check

		_ = a.TurnOff(device)
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
	var irw *bool = flag.Bool("irw", false, "infrared reading via LIRC")
	var irSimulatorKey *string = flag.String("ir-simulator", "", "simulate infrared events")

	flag.Parse()

	stopper := NewStopper()
	app := NewApplication(stopper.Add())

	harmonyHubAdapter := NewHarmonyHubAdapter("harmonyHubAdapter", "192.168.1.153:5222", stopper.Add())
	particleAdapter := NewParticleAdapter("particleAdapter", "310027000647343138333038", getParticleAccessToken())

	app.DefineAdapter(harmonyHubAdapter)
	app.DefineAdapter(particleAdapter)

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
