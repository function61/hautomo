package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/function61/eventhorizon/util/clicommon"
	"github.com/function61/home-automation-hub/adapters/eventghostnetworkclientadapter"
	"github.com/function61/home-automation-hub/adapters/happylightsadapter"
	"github.com/function61/home-automation-hub/adapters/harmonyhubadapter"
	"github.com/function61/home-automation-hub/adapters/ikeatradfriadapter"
	"github.com/function61/home-automation-hub/adapters/particleadapter"
	"github.com/function61/home-automation-hub/hapitypes"
	"github.com/function61/home-automation-hub/libraries/happylights/happylightsclientcli"
	"github.com/function61/home-automation-hub/libraries/happylights/happylightsserver"
	"github.com/function61/home-automation-hub/util/stopper"
	"github.com/function61/home-automation-hub/util/systemdinstaller"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"os"
)

// replaced in build process with actual version
var version = "dev"

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

func configureAppAndStartAdapters(app *Application, conf *hapitypes.ConfigFile, stop *stopper.Stopper) error {
	for _, adapter := range conf.Adapters {
		switch adapter.Type {
		case "particle":
			app.DefineAdapter(particleadapter.New(hapitypes.NewAdapter(adapter.Id), adapter))
		case "harmony":
			app.DefineAdapter(harmonyhubadapter.New(hapitypes.NewAdapter(adapter.Id), adapter, stop.Add()))
		case "ikea_tradfri":
			app.DefineAdapter(
				ikeatradfriadapter.New(hapitypes.NewAdapter(adapter.Id), adapter))
		case "happylights":
			app.DefineAdapter(happylightsadapter.New(hapitypes.NewAdapter(adapter.Id), adapter))
		case "eventghostnetworkclient":
			app.DefineAdapter(eventghostnetworkclientadapter.New(hapitypes.NewAdapter(adapter.Id), adapter, stop.Add()))
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
			return errors.New("unkown adapter: " + adapter.Type)
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

	return nil
}

var rootCmd = &cobra.Command{
	Use:   os.Args[0],
	Short: "Home automation hub from function61.com",
}

func startServer() {
	conf, confErr := readConfigurationFile()
	if confErr != nil {
		panic(confErr)
	}

	stop := stopper.New()
	app := NewApplication(stop.Add())

	if err := configureAppAndStartAdapters(app, conf, stop); err != nil {
		panic(err)
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

	if err := SyncToAlexaConnector(conf); err != nil {
		log.Printf("SyncToAlexaConnector: %s", err.Error())
	}

	clicommon.WaitForInterrupt()

	log.Println("main: received interrupt")

	stop.StopAll()

	log.Println("main: all components stopped")
}

func serverEntry() *cobra.Command {
	server := &cobra.Command{
		Use:   "server",
		Short: "Starts the server",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			startServer()
		},
	}

	server.AddCommand(&cobra.Command{
		Use:   "write-systemd-unit-file",
		Short: "Install unit file to start this on startup",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := systemdinstaller.InstallSystemdServiceFile("homeautomation", []string{"server"}, "home automation hub"); err != nil {
				panic(err)
			}
		},
	})

	return server
}

func main() {
	rootCmd.AddCommand(serverEntry())
	rootCmd.AddCommand(happylightsclientcli.BindEntrypoint(happylightsserver.Entrypoint()))
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Shows version number of this app",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s\n", version)
		},
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
