package main

import (
	"./util/systemdinstaller"
	"errors"
	"fmt"
	"github.com/function61/eventhorizon/util/clicommon"
	"log"
	"os"
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
					log.Printf("application: IR ignored: %s", ir.Event)
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
		adapter.PowerMsg <- NewPowerMsg(device.AdaptersDeviceId, device.PowerOnCmd, true)

		device.ProbablyTurnedOn = true
	} else if power.Kind == powerKindOff {
		log.Printf("Power off: %s", device.Name)

		adapter := a.adapterById[device.AdapterId]
		adapter.PowerMsg <- NewPowerMsg(device.AdaptersDeviceId, device.PowerOffCmd, false)

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
	if len(os.Args) == 2 && os.Args[1] == "--help" {
		fmt.Printf("Usage: %s [--write-systemd-unit-file]\n", os.Args[0])
		return
	}
	if len(os.Args) == 2 && os.Args[1] == "--write-systemd-unit-file" {
		if err := systemdinstaller.InstallSystemdServiceFile("homeautomation", "home automation hub"); err != nil {
			panic(err)
		}
		return
	}
	if len(os.Args) != 1 {
		fmt.Printf("Invalid arguments. Run %s --help", os.Args[0])
		os.Exit(1)
		return
	}

	conf, confErr := readConfigurationFile()
	if confErr != nil {
		panic(confErr)
	}

	stopper := NewStopper()
	app := NewApplication(stopper.Add())

	for _, adapter := range conf.Adapters {
		switch adapter.Type {
		case "particle":
			app.DefineAdapter(NewParticleAdapter(adapter.Id, adapter.ParticleId, adapter.ParticleAccessToken))
		case "harmony":
			app.DefineAdapter(NewHarmonyHubAdapter(adapter.Id, adapter.HarmonyAddr, stopper.Add()))
		case "happylights":
			app.DefineAdapter(NewHappylightsAdapter(adapter.Id, adapter.HappyLightsAddr))
		case "irsimulator":
			go infraredSimulator(app, adapter.IrSimulatorKey, stopper.Add())
		case "lirc":
			go irwPoller(app, stopper.Add())
		case "sqs":
			go sqsPollerLoop(app, adapter.SqsQueueUrl, adapter.SqsKeyId, adapter.SqsKeySecret, stopper.Add())
		default:
			panic(errors.New("unkown adapter: " + adapter.Type))
		}
	}

	for _, device := range conf.Devices {
		app.AttachDevice(NewDevice(
			device.DeviceId,
			device.AdapterId,
			device.AdaptersDeviceId,
			device.Name,
			device.Description,
			device.PowerOnCmd,
			device.PowerOffCmd))
	}

	for _, deviceGroup := range conf.DeviceGroups {
		app.AttachDeviceGroup(NewDeviceGroup(deviceGroup.Id, deviceGroup.Name, deviceGroup.DeviceIds))
	}

	app.SyncToCloud()

	clicommon.WaitForInterrupt()

	log.Println("main: received interrupt")

	stopper.StopAll()

	log.Println("main: all components stopped")
}
