package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/function61/eventhorizon/util/clicommon"
	"log"
	"strings"
)

type Application struct {
	adapterById           map[string]*Adapter
	deviceById            map[string]*Device
	deviceGroupById       map[string]*DeviceGroup
	infraredToPowerEvent  map[string]PowerEvent
	infraredToInfraredMsg map[string]InfraredToInfraredWrapper
	infraredEvent         chan InfraredEvent
	powerEvent            chan PowerEvent
}

type InfraredToInfraredWrapper struct {
	adapter     *Adapter
	infraredMsg InfraredMsg
}

func NewApplication(stopper *Stopper) *Application {
	app := &Application{
		adapterById:           make(map[string]*Adapter),
		deviceById:            make(map[string]*Device),
		deviceGroupById:       make(map[string]*DeviceGroup),
		infraredToPowerEvent:  make(map[string]PowerEvent),
		infraredToInfraredMsg: make(map[string]InfraredToInfraredWrapper),
		infraredEvent:         make(chan InfraredEvent, 1),
		powerEvent:            make(chan PowerEvent, 1),
	}

	go func() {
		defer stopper.Done()

		log.Println("application: started")

		for {
			select {
			case <-stopper.ShouldStop:
				log.Println("application: stopping")
				return
			case power := <-app.powerEvent:
				app.deviceOrDeviceGroupPower(power)
			case ir := <-app.infraredEvent:

				if powerEvent, ok := app.infraredToPowerEvent[ir.Event]; ok {
					log.Printf("application: IR: %s -> power for %s", ir.Event, powerEvent.DeviceIdOrDeviceGroupId)

					app.powerEvent <- powerEvent
				} else if i2i, ok := app.infraredToInfraredMsg[ir.Event]; ok {
					log.Printf("application: IR passthrough: %s -> %s", ir.Event, i2i.infraredMsg.Command)

					i2i.adapter.InfraredMsg <- i2i.infraredMsg
				} else {
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

func (a *Application) InfraredShouldPower(key string, powerEvent PowerEvent) {
	a.infraredToPowerEvent[key] = powerEvent
}

func (a *Application) InfraredShouldInfrared(key string, deviceId string, command string) {
	device := a.deviceById[deviceId]
	adapter := a.adapterById[device.AdapterId]

	a.infraredToInfraredMsg[key] = InfraredToInfraredWrapper{adapter, NewInfraredMsg(device.AdaptersDeviceId, command)}
}

func (a *Application) deviceOrDeviceGroupPower(power PowerEvent) error {
	device, deviceFound := a.deviceById[power.DeviceIdOrDeviceGroupId]
	if deviceFound {
		return a.devicePower(device, power)
	}

	deviceGroup, deviceGroupFound := a.deviceGroupById[power.DeviceIdOrDeviceGroupId]
	if deviceGroupFound {
		for _, deviceId := range deviceGroup.DeviceIds {
			device := a.deviceById[deviceId]

			_ = a.devicePower(device, power)
		}

		return nil
	}

	return errDeviceNotFound
}

func (a *Application) devicePower(device *Device, power PowerEvent) error {
	if power.Kind == powerKindOn {
		log.Printf("Power on: %s", device.Name)

		adapter := a.adapterById[device.AdapterId]
		adapter.PowerMsg <- NewPowerMsg(device.AdaptersDeviceId, device.PowerOnCmd)

		device.ProbablyTurnedOn = true
	} else if power.Kind == powerKindOff {
		log.Printf("Power off: %s", device.Name)

		adapter := a.adapterById[device.AdapterId]
		adapter.PowerMsg <- NewPowerMsg(device.AdaptersDeviceId, device.PowerOffCmd)

		device.ProbablyTurnedOn = false
	} else if power.Kind == powerKindToggle {
		log.Printf("Power toggle: %s, current state = %v", device.Name, device.ProbablyTurnedOn)

		if device.ProbablyTurnedOn {
			return a.devicePower(device, NewPowerEvent(device.Id, powerKindOff))
		} else {
			return a.devicePower(device, NewPowerEvent(device.Id, powerKindOn))
		}
	} else {
		panic(errors.New("unknown power kind"))
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
	app.AttachDevice(NewDevice("7e7453da", "harmonyHubAdapter", "47918441", "TV", `Philips 55" 4K 55PUS7909`, "PowerToggle", "PowerOff"))

	app.AttachDevice(NewDevice("d2ff0882", "particleAdapter", "", "Sofa light", "Floor light next the sofa", "C21", "C20"))
	app.AttachDevice(NewDevice("98d3cb01", "particleAdapter", "", "Speaker light", "Floor light under the speaker", "C31", "C30"))

	app.AttachDeviceGroup(NewDeviceGroup("cfb1b27f", "Living room lights", []string{
		"d2ff0882",
		"98d3cb01",
	}))

	/*
	app.InfraredShouldPower("KEY_VOLUMEUP", NewPowerEvent("98d3cb01", powerKindToggle))
	app.InfraredShouldPower("KEY_VOLUMEDOWN", NewPowerEvent("98d3cb01", powerKindOff))
	app.InfraredShouldPower("KEY_CHANNELUP", NewPowerEvent("d2ff0882", powerKindOn))
	app.InfraredShouldPower("KEY_CHANNELDOWN", NewPowerEvent("d2ff0882", powerKindOff))
	*/

	app.InfraredShouldInfrared("KEY_VOLUMEUP", "c0730bb2", "VolumeUp")
	app.InfraredShouldInfrared("KEY_VOLUMEDOWN", "c0730bb2", "VolumeDown")

	app.SyncToCloud()

	if *irw {
		go irwPoller(app, stopper.Add())
	}

	go sqsPollerLoop(app, stopper.Add())

	if *irSimulatorKey != "" {
		go infraredSimulator(app, *irSimulatorKey, stopper.Add())
	}

	clicommon.WaitForInterrupt()

	log.Println("main: received interrupt")

	stopper.StopAll()

	log.Println("main: all components stopped")
}
