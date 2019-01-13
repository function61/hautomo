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
	"github.com/function61/home-automation-hub/pkg/adapters/devicegroupadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/dummyadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/eventghostnetworkclientadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/harmonyhubadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/ikeatradfriadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/irsimulatoradapter"
	"github.com/function61/home-automation-hub/pkg/adapters/lircadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/particleadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/presencebypingadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/trionesadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/zigbee2mqttadapter"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

var log = logger.New("main")

// replaced in build process with actual version
var version = "dev"

type Application struct {
	adapterById   map[string]*hapitypes.Adapter
	deviceById    map[string]*hapitypes.Device
	subscriptions map[string]*hapitypes.SubscribeConfig
	inbound       *hapitypes.InboundFabric
}

type InfraredToInfraredWrapper struct {
	adapter     *hapitypes.Adapter
	infraredMsg hapitypes.InfraredMsg
}

func NewApplication(stop *stopper.Stopper) *Application {
	app := &Application{
		adapterById:   map[string]*hapitypes.Adapter{},
		deviceById:    map[string]*hapitypes.Device{},
		subscriptions: map[string]*hapitypes.SubscribeConfig{},
		inbound:       hapitypes.NewInboundFabric(),
	}

	go func() {
		defer stop.Done()

		log.Info(fmt.Sprintf("home-automation-hub %s started", version))
		defer log.Info("stopped")

		for {
			select {
			case <-stop.Signal:
				return
			case genericEvent := <-app.inbound.Ch:
				switch e := genericEvent.(type) {
				case *hapitypes.PersonPresenceChangeEvent:
					log.Info(fmt.Sprintf(
						"Person %s presence changed to %v",
						e.PersonId,
						e.Present))
				case *hapitypes.PowerEvent:
					device := app.deviceById[e.DeviceIdOrDeviceGroupId]

					if err := app.devicePower(device, e); err != nil {
						log.Error(err.Error())
					}
				case *hapitypes.ColorTemperatureEvent:
					device := app.deviceById[e.Device]
					adapter := app.adapterById[device.Conf.AdapterId]

					adapter.Send(hapitypes.NewColorTemperatureEvent(
						device.Conf.AdaptersDeviceId,
						e.TemperatureInKelvin))
				case *hapitypes.ColorMsg:
					device := app.deviceById[e.DeviceId]
					adapter := app.adapterById[device.Conf.AdapterId]

					device.LastColor = e.Color

					adapter.Send(hapitypes.NewColorMsg(
						device.Conf.AdaptersDeviceId,
						e.Color))
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

					adapter.Send(hapitypes.NewPlaybackEvent(
						device.Conf.AdaptersDeviceId,
						e.Action))
				case *hapitypes.BlinkEvent:
					device := app.deviceById[e.DeviceId]
					adapter := app.adapterById[device.Conf.AdapterId]

					adapter.Send(hapitypes.NewBlinkEvent(device.Conf.AdaptersDeviceId))
				case *hapitypes.InfraredEvent:
					app.publish(fmt.Sprintf("infrared:%s:%s", e.Remote, e.Event))
				case *hapitypes.PublishEvent:
					app.publish(e.Event)
				default:
					log.Error("Unsupported inbound event: " + genericEvent.InboundEventType())
				}
			}
		}
	}()

	return app
}

func (a *Application) publish(event string) {
	subscription, found := a.subscriptions[event]
	if !found {
		log.Debug(fmt.Sprintf("event %s ignored", event))
		return
	} else {
		log.Debug(fmt.Sprintf("event %s", event))
	}

	for _, action := range subscription.Actions {
		switch action.Verb {
		case "powerOn":
			a.inbound.Receive(hapitypes.NewPowerEvent(action.Device, hapitypes.PowerKindOn))
		case "powerOff":
			a.inbound.Receive(hapitypes.NewPowerEvent(action.Device, hapitypes.PowerKindOff))
		case "powerToggle":
			a.inbound.Receive(hapitypes.NewPowerEvent(action.Device, hapitypes.PowerKindToggle))
		case "blink":
			a.inbound.Receive(hapitypes.NewBlinkEvent(action.Device))
		case "ir":
			device := a.deviceById[action.Device]
			adapter := a.adapterById[device.Conf.AdapterId]

			msg := hapitypes.NewInfraredMsg(action.Device, action.IrCommand)
			adapter.Send(&msg)
		default:
			panic("unknown verb: " + action.Verb)
		}
	}
}

func (a *Application) devicePower(device *hapitypes.Device, power *hapitypes.PowerEvent) error {
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

type AdapterInitFn func(adapter *hapitypes.Adapter, stop *stopper.Stopper) error

func configureAppAndStartAdapters(app *Application, conf *hapitypes.ConfigFile, stopManager *stopper.Manager) error {
	adapters := map[string]AdapterInitFn{
		"devicegroup":             devicegroupadapter.Start,
		"dummy":                   dummyadapter.Start,
		"eventghostnetworkclient": eventghostnetworkclientadapter.Start,
		"triones":                 trionesadapter.Start,
		"harmony":                 harmonyhubadapter.Start,
		"ikea_tradfri":            ikeatradfriadapter.Start,
		"zigbee2mqtt":             zigbee2mqttadapter.Start,
		"irsimulator":             irsimulatoradapter.Start,
		"lirc":                    lircadapter.Start,
		"particle":                particleadapter.Start,
		"presencebyping":          presencebypingadapter.Start,
		"sqs":                     alexaadapter.Start,
	}

	for _, devGroup := range conf.DeviceGroups {
		generatedAdapterId := devGroup.DeviceId + "Group"

		adapterConf := hapitypes.AdapterConfig{
			Id:                 generatedAdapterId,
			Type:               "devicegroup",
			DevicegroupDevices: devGroup.Devices,
		}

		firstDeviceOfGroup := findDeviceConfig(devGroup.Devices[0], conf)
		if firstDeviceOfGroup == nil {
			return fmt.Errorf("device group device not found: %s", devGroup.Devices[0])
		}

		deviceConf := hapitypes.DeviceConfig{
			DeviceId:      devGroup.DeviceId,
			AdapterId:     adapterConf.Id,
			Name:          devGroup.Name,
			Description:   "Device group",
			AlexaCategory: firstDeviceOfGroup.AlexaCategory,
			Type:          firstDeviceOfGroup.Type, // TODO: compute lowest common denominator type?
		}

		conf.Adapters = append(conf.Adapters, adapterConf)
		conf.Devices = append(conf.Devices, deviceConf)
	}

	for _, adapterConf := range conf.Adapters {
		initFn, ok := adapters[adapterConf.Type]
		if !ok {
			return errors.New("unkown adapter: " + adapterConf.Type)
		}

		adapter := hapitypes.NewAdapter(adapterConf, conf, app.inbound)

		if err := initFn(adapter, stopManager.Stopper()); err != nil {
			return err
		}

		app.adapterById[adapter.Conf.Id] = adapter
	}

	for _, deviceConf := range conf.Devices {
		if _, exists := app.deviceById[deviceConf.DeviceId]; exists {
			return fmt.Errorf("duplicate device id %s", deviceConf.DeviceId)
		}

		device := hapitypes.NewDevice(deviceConf)
		app.deviceById[deviceConf.DeviceId] = device
	}

	for _, subscription := range conf.Subscriptions {
		// FIXME: how to do this better?
		tmp := subscription
		app.subscriptions[subscription.Event] = &tmp
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

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Error(fmt.Sprintf("ListenAndServe(): %s", err.Error()))
	}
}

func runServer() error {
	conf, confErr := readConfigurationFile()
	if confErr != nil {
		return confErr
	}

	stopManager := stopper.NewManager()
	defer log.Info("all components stopped")
	defer stopManager.StopAllWorkersAndWait()

	// FIXME: main loop probably shouldn't start here, since there's a race condition
	app := NewApplication(stopManager.Stopper())

	if err := configureAppAndStartAdapters(app, conf, stopManager); err != nil {
		return err
	}

	go handleHttp(conf, stopManager.Stopper())

	log.Info(fmt.Sprintf("stopping due to signal %s", ossignal.WaitForInterruptOrTerminate()))

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

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
