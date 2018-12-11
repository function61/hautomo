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
	"github.com/function61/home-automation-hub/pkg/adapters/devicegroupadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/dummyadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/eventghostnetworkclientadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/happylightsadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/harmonyhubadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/ikeatradfriadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/irsimulatoradapter"
	"github.com/function61/home-automation-hub/pkg/adapters/lircadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/particleadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/presencebypingadapter"
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
		infraredToPowerEvent:  make(map[string]hapitypes.PowerEvent),
		infraredToInfraredMsg: make(map[string]InfraredToInfraredWrapper),
		fabric:                signalfabric.New(),
	}

	go func() {
		defer stop.Done()

		log.Info(fmt.Sprintf("home-automation-hub %s started", version))
		defer log.Info("stopped")

		fabric := app.fabric

		for {
			select {
			case <-stop.Signal:
				return
			case genericEvent := <-fabric.Event:
				switch e := genericEvent.(type) {
				case *hapitypes.PersonPresenceChangeEvent:
					log.Info(fmt.Sprintf(
						"Person %s presence changed to %v",
						e.PersonId,
						e.Present))
				case *hapitypes.PowerEvent:
					device := app.deviceById[e.DeviceIdOrDeviceGroupId]

					if err := app.devicePower(device, *e); err != nil {
						log.Error(err.Error())
					}
				case *hapitypes.ColorTemperatureEvent:
					device := app.deviceById[e.Device]
					adapter := app.adapterById[device.Conf.AdapterId]

					e2 := hapitypes.NewColorTemperatureEvent(
						device.Conf.AdaptersDeviceId,
						e.TemperatureInKelvin)
					adapter.Send(&e2)
				case *hapitypes.ColorMsg:
					device := app.deviceById[e.DeviceId]
					adapter := app.adapterById[device.Conf.AdapterId]

					device.LastColor = e.Color

					adaptedColorMsg := hapitypes.NewColorMsg(device.Conf.AdaptersDeviceId, e.Color)
					adapter.Send(&adaptedColorMsg)
				case *hapitypes.BrightnessEvent:
					device := app.deviceById[e.DeviceIdOrDeviceGroupId]
					adapter := app.adapterById[device.Conf.AdapterId]

					e2 := hapitypes.NewBrightnessMsg(
						device.Conf.AdaptersDeviceId,
						e.Brightness,
						device.LastColor)
					adapter.Send(&e2)
				case *hapitypes.PlaybackEvent:
					device := app.deviceById[e.DeviceIdOrDeviceGroupId]
					adapter := app.adapterById[device.Conf.AdapterId]

					e2 := hapitypes.NewPlaybackEvent(device.Conf.AdaptersDeviceId, e.Action)
					adapter.Send(&e2)
				case *hapitypes.InfraredEvent:
					if powerEvent, ok := app.infraredToPowerEvent[e.Event]; ok {
						log.Debug(fmt.Sprintf("IR: %s -> power for %s", e.Event, powerEvent.DeviceIdOrDeviceGroupId))

						fabric.Receive(&powerEvent)
					} else if i2i, ok := app.infraredToInfraredMsg[e.Event]; ok {
						log.Debug(fmt.Sprintf("IR passthrough: %s -> %s", e.Event, i2i.infraredMsg.Command))

						i2i.adapter.Send(&i2i.infraredMsg)
					} else {
						log.Debug(fmt.Sprintf("IR ignored: %s", e.Event))
					}
				default:
					log.Error("Unsupported inbound event: " + genericEvent.InboundEventType())
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
	a.deviceById[device.Conf.DeviceId] = device
}

func (a *Application) InfraredShouldPower(key string, powerEvent hapitypes.PowerEvent) {
	a.infraredToPowerEvent[key] = powerEvent
}

func (a *Application) InfraredShouldInfrared(key string, deviceId string, command string) {
	device := a.deviceById[deviceId]
	adapter := a.adapterById[device.Conf.AdapterId]

	msg := hapitypes.NewInfraredMsg(device.Conf.AdaptersDeviceId, command)
	a.infraredToInfraredMsg[key] = InfraredToInfraredWrapper{adapter, msg}
}

func (a *Application) devicePower(device *hapitypes.Device, power hapitypes.PowerEvent) error {
	if power.Kind == hapitypes.PowerKindOn {
		log.Debug(fmt.Sprintf("Power on: %s", device.Conf.Name))

		adapter := a.adapterById[device.Conf.AdapterId]
		e := hapitypes.NewPowerMsg(device.Conf.AdaptersDeviceId, device.Conf.PowerOnCmd, true)
		adapter.Send(&e)

		device.ProbablyTurnedOn = true
	} else if power.Kind == hapitypes.PowerKindOff {
		log.Debug(fmt.Sprintf("Power off: %s", device.Conf.Name))

		adapter := a.adapterById[device.Conf.AdapterId]
		e := hapitypes.NewPowerMsg(device.Conf.AdaptersDeviceId, device.Conf.PowerOffCmd, false)
		adapter.Send(&e)

		device.ProbablyTurnedOn = false
	} else if power.Kind == hapitypes.PowerKindToggle {
		log.Debug(fmt.Sprintf("Power toggle: %s, current state = %v", device.Conf.Name, device.ProbablyTurnedOn))

		if device.ProbablyTurnedOn {
			return a.devicePower(device, hapitypes.NewPowerEvent(device.Conf.DeviceId, hapitypes.PowerKindOff))
		} else {
			return a.devicePower(device, hapitypes.NewPowerEvent(device.Conf.DeviceId, hapitypes.PowerKindOn))
		}
	} else {
		return errors.New("unknown power kind")
	}

	return nil
}

func configureAppAndStartAdapters(app *Application, conf *hapitypes.ConfigFile, stopManager *stopper.Manager) error {
	// TODO: map[string]InitFn

	for _, adapterConf := range conf.Adapters {
		adapter := hapitypes.NewAdapter(adapterConf.Id)

		switch adapterConf.Type {
		case "devicegroup":
			devicegroupadapter.New(adapter, app.fabric, adapterConf)
			app.DefineAdapter(adapter)
		case "particle":
			particleadapter.New(adapter, adapterConf)
			app.DefineAdapter(adapter)
		case "presencebyping":
			presencebypingadapter.StartSensor(adapterConf, app.fabric, stopManager.Stopper())
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
		case "dummy":
			dummyadapter.New(adapter, adapterConf)
			app.DefineAdapter(adapter)
		default:
			return errors.New("unkown adapter: " + adapterConf.Type)
		}
	}

	for _, deviceConf := range conf.Devices {
		app.AttachDevice(hapitypes.NewDevice(deviceConf))
	}

	supportedIrPowerKinds := map[string]hapitypes.PowerKind{
		"toggle": hapitypes.PowerKindToggle,
		"on":     hapitypes.PowerKindOn,
		"off":    hapitypes.PowerKindOff,
	}

	for _, powerConfig := range conf.IrPowers {
		kind, ok := supportedIrPowerKinds[powerConfig.PowerKind]
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
	log := logger.New("handleHttp")

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
		Use:   "lint",
		Short: "Verifies the syntax of the configuration file",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			_, err := readConfigurationFile()
			if err != nil {
				panic(err)
			}
		},
	})

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
