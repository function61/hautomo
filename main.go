package main

import (
	"./adapters/eventghostnetworkclientadapter"
	"./adapters/happylightsadapter"
	"./adapters/harmonyhubadapter"
	"./adapters/particleadapter"
	"./hapitypes"
	"./util/stopper"
	"./util/systemdinstaller"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/function61/eventhorizon/util/clicommon"
	"log"
	"net/http"
	"os"
	"strings"
)

type Application struct {
	adapterById           map[string]*hapitypes.Adapter
	deviceById            map[string]*hapitypes.Device
	deviceGroupById       map[string]*hapitypes.DeviceGroup
	infraredToPowerEvent  map[string]hapitypes.PowerEvent
	infraredToInfraredMsg map[string]InfraredToInfraredWrapper
	infraredEvent         chan hapitypes.InfraredEvent
	powerEvent            chan hapitypes.PowerEvent
	colorEvent            chan hapitypes.ColorMsg
	brightnessEvent       chan hapitypes.BrightnessEvent
	playbackEvent         chan hapitypes.PlaybackEvent
}

type InfraredToInfraredWrapper struct {
	adapter     *hapitypes.Adapter
	infraredMsg hapitypes.InfraredMsg
}

func NewApplication(stop *stopper.Stopper) *Application {
	app := &Application{
		adapterById:           make(map[string]*hapitypes.Adapter),
		deviceById:            make(map[string]*hapitypes.Device),
		deviceGroupById:       make(map[string]*hapitypes.DeviceGroup),
		infraredToPowerEvent:  make(map[string]hapitypes.PowerEvent),
		infraredToInfraredMsg: make(map[string]InfraredToInfraredWrapper),
		infraredEvent:         make(chan hapitypes.InfraredEvent, 1),
		powerEvent:            make(chan hapitypes.PowerEvent, 1),
		colorEvent:            make(chan hapitypes.ColorMsg, 1),
		brightnessEvent:       make(chan hapitypes.BrightnessEvent, 1),
		playbackEvent:         make(chan hapitypes.PlaybackEvent, 1),
	}

	go func() {
		defer stop.Done()

		log.Println("application: started")

		for {
			select {
			case <-stop.ShouldStop:
				log.Println("application: stopping")
				return
			case power := <-app.powerEvent:
				app.deviceOrDeviceGroupPower(power)
			case colorMsg := <-app.colorEvent:
				// TODO: device group support
				device := app.deviceById[colorMsg.DeviceId]
				adapter := app.adapterById[device.AdapterId]

				device.LastColor = colorMsg.Color

				adaptedColorMsg := hapitypes.NewColorMsg(device.AdaptersDeviceId, colorMsg.Color)

				adapter.ColorMsg <- adaptedColorMsg
			case brightnessEvent := <-app.brightnessEvent:
				// TODO: device group support
				device := app.deviceById[brightnessEvent.DeviceIdOrDeviceGroupId]
				adapter := app.adapterById[device.AdapterId]

				dimmedColor := hapitypes.RGB{
					Red:   uint8(float64(device.LastColor.Red) * float64(brightnessEvent.Brightness) / 100.0),
					Green: uint8(float64(device.LastColor.Green) * float64(brightnessEvent.Brightness) / 100.0),
					Blue:  uint8(float64(device.LastColor.Blue) * float64(brightnessEvent.Brightness) / 100.0),
				}

				adapter.ColorMsg <- hapitypes.NewColorMsg(device.AdaptersDeviceId, dimmedColor)
			case playbackEvent := <-app.playbackEvent:
				// TODO: device group support
				device := app.deviceById[playbackEvent.DeviceIdOrDeviceGroupId]
				adapter := app.adapterById[device.AdapterId]

				adapter.PlaybackMsg <- hapitypes.NewPlaybackEvent(device.AdaptersDeviceId, playbackEvent.Action)
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

func (a *Application) DefineAdapter(adapter *hapitypes.Adapter) {
	a.adapterById[adapter.Id] = adapter
}

func (a *Application) AttachDevice(device *hapitypes.Device) {
	a.deviceById[device.Id] = device
}

func (a *Application) AttachDeviceGroup(deviceGroup *hapitypes.DeviceGroup) {
	a.deviceGroupById[deviceGroup.Id] = deviceGroup
}

func (a *Application) InfraredShouldPower(key string, powerEvent hapitypes.PowerEvent) {
	a.infraredToPowerEvent[key] = powerEvent
}

func (a *Application) InfraredShouldInfrared(key string, deviceId string, command string) {
	device := a.deviceById[deviceId]
	adapter := a.adapterById[device.AdapterId]

	a.infraredToInfraredMsg[key] = InfraredToInfraredWrapper{adapter, hapitypes.NewInfraredMsg(device.AdaptersDeviceId, command)}
}

func (a *Application) deviceOrDeviceGroupPower(power hapitypes.PowerEvent) error {
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

	return hapitypes.ErrDeviceNotFound
}

func (a *Application) devicePower(device *hapitypes.Device, power hapitypes.PowerEvent) error {
	if power.Kind == hapitypes.PowerKindOn {
		log.Printf("Power on: %s", device.Name)

		adapter := a.adapterById[device.AdapterId]
		adapter.PowerMsg <- hapitypes.NewPowerMsg(device.AdaptersDeviceId, device.PowerOnCmd, true)

		device.ProbablyTurnedOn = true
	} else if power.Kind == hapitypes.PowerKindOff {
		log.Printf("Power off: %s", device.Name)

		adapter := a.adapterById[device.AdapterId]
		adapter.PowerMsg <- hapitypes.NewPowerMsg(device.AdaptersDeviceId, device.PowerOffCmd, false)

		device.ProbablyTurnedOn = false
	} else if power.Kind == hapitypes.PowerKindToggle {
		log.Printf("Power toggle: %s, current state = %v", device.Name, device.ProbablyTurnedOn)

		if device.ProbablyTurnedOn {
			return a.devicePower(device, hapitypes.NewPowerEvent(device.Id, hapitypes.PowerKindOff))
		} else {
			return a.devicePower(device, hapitypes.NewPowerEvent(device.Id, hapitypes.PowerKindOn))
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

	stop := stopper.New()
	app := NewApplication(stop.Add())

	for _, adapter := range conf.Adapters {
		switch adapter.Type {
		case "particle":
			app.DefineAdapter(particleadapter.New(
				adapter.Id,
				adapter.ParticleId,
				adapter.ParticleAccessToken))
		case "harmony":
			app.DefineAdapter(harmonyhubadapter.New(
				adapter.Id,
				adapter.HarmonyAddr,
				stop.Add()))
		case "happylights":
			app.DefineAdapter(happylightsadapter.New(
				adapter.Id,
				adapter.HappyLightsAddr))
		case "eventghostnetworkclient":
			app.DefineAdapter(eventghostnetworkclientadapter.New(
				adapter.Id,
				adapter.EventghostAddr,
				adapter.EventghostSecret,
				stop.Add()))
		case "irsimulator":
			go infraredSimulator(
				app,
				adapter.IrSimulatorKey,
				stop.Add())
		case "lirc":
			go irwPoller(
				app,
				stop.Add())
		case "sqs":
			go sqsPollerLoop(
				app,
				adapter.SqsQueueUrl,
				adapter.SqsKeyId,
				adapter.SqsKeySecret,
				stop.Add())
		default:
			panic(errors.New("unkown adapter: " + adapter.Type))
		}
	}

	for _, device := range conf.Devices {
		app.AttachDevice(hapitypes.NewDevice(
			device.DeviceId,
			device.AdapterId,
			device.AdaptersDeviceId,
			device.Name,
			device.Description,
			device.PowerOnCmd,
			device.PowerOffCmd))
	}

	for _, deviceGroup := range conf.DeviceGroups {
		app.AttachDeviceGroup(hapitypes.NewDeviceGroup(deviceGroup.Id, deviceGroup.Name, deviceGroup.DeviceIds))
	}

	supportedPowerKinds := map[string]hapitypes.PowerKind{
		"toggle": hapitypes.PowerKindToggle,
		"on":     hapitypes.PowerKindOn,
		"off":    hapitypes.PowerKindOff,
	}

	for _, powerConfig := range conf.IrPowers {
		kind, ok := supportedPowerKinds[powerConfig.PowerKind]
		if !ok {
			panic(fmt.Errorf("Unsupported power kind: %s", powerConfig.PowerKind))
		}

		app.InfraredShouldPower(powerConfig.RemoteKey, hapitypes.NewPowerEvent(powerConfig.ToDevice, kind))
	}

	for _, ir2ir := range conf.IrToIr {
		app.InfraredShouldInfrared(ir2ir.RemoteKey, ir2ir.ToDevice, ir2ir.IrEvent)
	}

	go func(stop *stopper.Stopper) {
		defer stop.Done()
		srv := &http.Server{Addr: ":8080"}

		go func() {
			<-stop.ShouldStop

			log.Printf("httpserver: requesting stop")

			_ = srv.Shutdown(nil)
		}()

		http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
			enc := json.NewEncoder(w)
			enc.SetIndent("", "  ")
			enc.Encode(conf)
		})

		if err := srv.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			log.Printf("httpserver: stopped because: %s", err)
		}
	}(stop.Add())

	app.SyncToCloud()

	clicommon.WaitForInterrupt()

	log.Println("main: received interrupt")

	stop.StopAll()

	log.Println("main: all components stopped")
}
