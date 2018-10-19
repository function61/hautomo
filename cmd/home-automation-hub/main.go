package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/function61/gokit/logger"
	"github.com/function61/gokit/ossignal"
	"github.com/function61/gokit/stopper"
	"github.com/function61/gokit/systemdinstaller"
	"github.com/function61/home-automation-hub/pkg/adapters/alexaadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/alexaadapter/alexadevicesync"
	"github.com/function61/home-automation-hub/pkg/adapters/eventghostnetworkclientadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/happylightsadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/harmonyhubadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/ikeatradfriadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/irsimulatoradapter"
	"github.com/function61/home-automation-hub/pkg/adapters/lircadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/particleadapter"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
	"github.com/function61/home-automation-hub/pkg/happylights/happylightsclientcli"
	"github.com/function61/home-automation-hub/pkg/happylights/happylightsserver"
	"github.com/function61/home-automation-hub/pkg/signalfabric"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

var log = logger.New("main")

// replaced in build process with actual version
var version = "dev"

type Application struct {
	adapterById           map[string]*hapitypes.Adapter
	deviceById            map[string]*hapitypes.Device
	deviceGroupById       map[string]*hapitypes.DeviceGroup
	infraredToPowerEvent  map[string]hapitypes.PowerEvent
	infraredToInfraredMsg map[string]InfraredToInfraredWrapper
	fabric                *signalfabric.Fabric
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
		fabric:                signalfabric.New(),
	}

	go func() {
		defer stop.Done()

		log.Info("started")
		defer log.Info("stopped")

		fabric := app.fabric

		for {
			select {
			case <-stop.Signal:
				return
			case power := <-fabric.PowerEvent:
				app.deviceOrDeviceGroupPower(power)
			case colorMsg := <-fabric.ColorEvent:
				// TODO: device group support
				device := app.deviceById[colorMsg.DeviceId]
				adapter := app.adapterById[device.AdapterId]

				device.LastColor = colorMsg.Color

				adaptedColorMsg := hapitypes.NewColorMsg(device.AdaptersDeviceId, colorMsg.Color)

				adapter.ColorMsg <- adaptedColorMsg
			case brightnessEvent := <-fabric.BrightnessEvent:
				// TODO: device group support
				device := app.deviceById[brightnessEvent.DeviceIdOrDeviceGroupId]
				adapter := app.adapterById[device.AdapterId]

				adapter.BrightnessMsg <- hapitypes.NewBrightnessMsg(
					device.AdaptersDeviceId,
					brightnessEvent.Brightness,
					device.LastColor)
			case playbackEvent := <-fabric.PlaybackEvent:
				// TODO: device group support
				device := app.deviceById[playbackEvent.DeviceIdOrDeviceGroupId]
				adapter := app.adapterById[device.AdapterId]

				adapter.PlaybackMsg <- hapitypes.NewPlaybackEvent(device.AdaptersDeviceId, playbackEvent.Action)
			case ir := <-fabric.InfraredEvent:
				if powerEvent, ok := app.infraredToPowerEvent[ir.Event]; ok {
					log.Debug(fmt.Sprintf("IR: %s -> power for %s", ir.Event, powerEvent.DeviceIdOrDeviceGroupId))

					fabric.PowerEvent <- powerEvent
				} else if i2i, ok := app.infraredToInfraredMsg[ir.Event]; ok {
					log.Debug(fmt.Sprintf("IR passthrough: %s -> %s", ir.Event, i2i.infraredMsg.Command))

					i2i.adapter.InfraredMsg <- i2i.infraredMsg
				} else {
					log.Debug(fmt.Sprintf("IR ignored: %s", ir.Event))
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
		log.Debug(fmt.Sprintf("Power on: %s", device.Name))

		adapter := a.adapterById[device.AdapterId]
		adapter.PowerMsg <- hapitypes.NewPowerMsg(device.AdaptersDeviceId, device.PowerOnCmd, true)

		device.ProbablyTurnedOn = true
	} else if power.Kind == hapitypes.PowerKindOff {
		log.Debug(fmt.Sprintf("Power off: %s", device.Name))

		adapter := a.adapterById[device.AdapterId]
		adapter.PowerMsg <- hapitypes.NewPowerMsg(device.AdaptersDeviceId, device.PowerOffCmd, false)

		device.ProbablyTurnedOn = false
	} else if power.Kind == hapitypes.PowerKindToggle {
		log.Debug(fmt.Sprintf("Power toggle: %s, current state = %v", device.Name, device.ProbablyTurnedOn))

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

func configureAppAndStartAdapters(app *Application, conf *hapitypes.ConfigFile, stopManager *stopper.Manager) error {
	// TODO: map[string]InitFn

	for _, adapterConf := range conf.Adapters {
		adapter := hapitypes.NewAdapter(adapterConf.Id)

		switch adapterConf.Type {
		case "particle":
			particleadapter.New(adapter, adapterConf)
			app.DefineAdapter(adapter)
		case "harmony":
			harmonyhubadapter.New(adapter, adapterConf, stopManager.Stopper())
			app.DefineAdapter(adapter)
		case "ikea_tradfri":
			ikeatradfriadapter.New(adapter, adapterConf)
			app.DefineAdapter(adapter)
		case "happylights":
			happylightsadapter.New(adapter, adapterConf)
			app.DefineAdapter(adapter)
		case "eventghostnetworkclient":
			eventghostnetworkclientadapter.New(adapter, adapterConf, stopManager.Stopper())
			app.DefineAdapter(adapter)
		case "irsimulator":
			irsimulatoradapter.StartSensor(adapter, adapterConf, app.fabric, stopManager.Stopper())
			app.DefineAdapter(adapter)
		case "lirc":
			go lircadapter.StartSensor(
				app.fabric,
				stopManager.Stopper())
			// FIXME: app.DefineAdapter() intentionally not called
		case "sqs":
			go alexaadapter.StartSensor(
				app.fabric,
				adapterConf,
				stopManager.Stopper())
			// FIXME: app.DefineAdapter() intentionally not called
		default:
			return errors.New("unkown adapter: " + adapterConf.Type)
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

func handleHttp(conf *hapitypes.ConfigFile, stop *stopper.Stopper) {
	defer stop.Done()
	srv := &http.Server{Addr: ":8080"}

	go func() {
		<-stop.Signal

		log.Info("stopping HTTP")

		_ = srv.Shutdown(nil)
	}()

	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.Encode(conf)
	})

	if err := srv.ListenAndServe(); err != nil {
		// cannot panic, because this probably is an intentional close
		log.Error(fmt.Sprintf("ListenAndServe() stopped: %s", err.Error()))
	}
}

func runServer() error {
	conf, confErr := readConfigurationFile()
	if confErr != nil {
		return confErr
	}

	stopManager := stopper.NewManager()

	app := NewApplication(stopManager.Stopper())

	if err := configureAppAndStartAdapters(app, conf, stopManager); err != nil {
		return err
	}

	go handleHttp(conf, stopManager.Stopper())

	if err := alexadevicesync.Sync(conf); err != nil {
		log.Error(fmt.Sprintf("alexadevicesync: %s", err.Error()))
	}

	log.Info("alexadevicesync completed")

	log.Info(fmt.Sprintf("stopping due to signal %s", ossignal.WaitForInterruptOrTerminate()))

	stopManager.StopAllWorkersAndWait()

	log.Info("all components stopped")

	return nil
}

func serverEntry() *cobra.Command {
	server := &cobra.Command{
		Use:   "server",
		Short: "Starts the server",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := runServer(); err != nil {
				panic(err)
			}
		},
	}

	server.AddCommand(&cobra.Command{
		Use:   "write-systemd-unit-file",
		Short: "Install unit file to start this on startup",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			systemdHints, err := systemdinstaller.InstallSystemdServiceFile("homeautomation", []string{"server"}, "home automation hub")
			if err != nil {
				panic(err)
			}

			fmt.Println(systemdHints)
		},
	})

	return server
}

func main() {
	rootCmd := &cobra.Command{
		Use:     os.Args[0],
		Short:   "Home automation hub from function61.com",
		Version: version,
	}
	rootCmd.AddCommand(serverEntry())
	rootCmd.AddCommand(happylightsclientcli.BindEntrypoint(happylightsserver.Entrypoint()))

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
