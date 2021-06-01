// Easy Zigbee Stack - the goal is to be easiest-to-understand Zigbee codebase available.
package ezstack

import (
	"context"
	"errors"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/function61/gokit/log/logex"
	"github.com/function61/gokit/sync/taskrunner"
	"github.com/function61/hautomo/pkg/ezstack/binstruct"
	"github.com/function61/hautomo/pkg/ezstack/coordinator"
	"github.com/function61/hautomo/pkg/ezstack/zcl"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
	"github.com/function61/hautomo/pkg/ezstack/zcl/frame"
	"github.com/function61/hautomo/pkg/ezstack/znp"
	"github.com/function61/hautomo/pkg/ezstack/znp/unp"
	"go.bug.st/serial"
)

const (
	DefaultSingleEndpointId = 1 // for simple single-endpoint devices, its endpoint ID usually is 1
)

// FIXME: these are all bad
var (
	logger = logex.StandardLogger()
	log    = logex.Prefix("ezstack", logger)
	logl   = logex.Levels(log)
)

type Channels struct {
	onDeviceRegistered      chan *Device
	onDeviceUnregistered    chan *Device
	onDeviceBecameAvailable chan *Device
	onDeviceIncomingMessage chan *DeviceIncomingMessage
}

func (c *Channels) OnDeviceRegistered() chan *Device {
	return c.onDeviceRegistered
}

// TODO: document what's the difference between available and registered
// seems to be signalled only when device's network address changes
func (c *Channels) OnDeviceBecameAvailable() chan *Device {
	return c.onDeviceBecameAvailable
}

func (c *Channels) OnDeviceUnregistered() chan *Device {
	return c.onDeviceUnregistered
}

// "application-level" message, i.e. sensor sending data
// TODO: rename to reduce confusion between device registration (name sounds like device is incoming to the cluster..)
func (c *Channels) OnDeviceIncomingMessage() chan *DeviceIncomingMessage {
	return c.onDeviceIncomingMessage
}

type NodeDatabase interface {
	InsertDevice(*Device) error
	GetDevice(ieeeAddress string) (*Device, bool)
	GetDeviceByNetworkAddress(nwkAddress string) (*Device, bool)
	RemoveDevice(ieeeAddress string) error
}

type Stack struct {
	db                NodeDatabase
	configuration     coordinator.Configuration
	coordinator       *coordinator.Coordinator
	registrationQueue chan *znp.ZdoEndDeviceAnnceInd
	zcl               *zcl.Zcl
	channels          *Channels
}

func New(configuration coordinator.Configuration, db NodeDatabase) *Stack {
	coordinator := coordinator.New(&configuration)

	zcl := zcl.Library

	return &Stack{
		db:                db,
		configuration:     configuration,
		coordinator:       coordinator,
		registrationQueue: make(chan *znp.ZdoEndDeviceAnnceInd),
		zcl:               zcl,
		channels: &Channels{
			onDeviceRegistered:      make(chan *Device, 10),
			onDeviceBecameAvailable: make(chan *Device, 10),
			onDeviceUnregistered:    make(chan *Device, 10),
			onDeviceIncomingMessage: make(chan *DeviceIncomingMessage, 100),
		},
	}
}

// if *packetCaptureFile* non-empty, specifies a file to log inbound UNP frames
func (s *Stack) Run(ctx context.Context, joinEnable bool, packetCaptureFilename string, settingsFlash bool) error {
	logl.Debug.Printf(
		"opening Zigbee radio %s at %d bauds/s",
		s.configuration.Serial.Port,
		s.configuration.Serial.BaudRateOrDefault())

	port, err := openPort(
		s.configuration.Serial.Port,
		s.configuration.Serial.BaudRateOrDefault())
	if err != nil {
		return fmt.Errorf("openPort: %s: %w", s.configuration.Serial.Port, err)
	}
	defer port.Close()

	// connect to ZNP using UNP protocol with serial port as a transport
	networkProcessor := znp.New(unp.NewWith8BitsPayloadLength(port), logex.Prefix("znp", logger))

	tasks := taskrunner.New(ctx, log)

	if packetCaptureFilename != "" {
		tasks.Start("packetcapture", func(ctx context.Context) error {
			return runPacketCapture(ctx, packetCaptureFilename, networkProcessor)
		})
	}

	// multiple ways for us to need port closing, so this is mainly a hack
	tasks.Start("portcloser", func(ctx context.Context) error {
		<-ctx.Done()

		// ZNP is most likely blocking on an UNP read
		return port.Close() // double close intentional
	})

	tasks.Start("znp", func(ctx context.Context) error {
		return networkProcessor.Run(ctx)
	})

	tasks.Start("coordinator", func(ctx context.Context) error {
		return s.coordinator.Run(ctx, joinEnable, networkProcessor, settingsFlash)
	})

	// to have expensive operation in separate non-blocking thread, but still do multiple registrations
	// sequentially
	tasks.Start("registrationqueue", func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case announcedDevice, ok := <-s.registrationQueue:
				if !ok { // queue closed
					return nil
				}

				if err := s.registerDevice(announcedDevice); err != nil {
					logl.Error.Printf("registerDevice: %s", err.Error())
				}
			}
		}
	})

	for {
		select {
		case <-ctx.Done():
			return tasks.Wait()
		case err := <-tasks.Done():
			return err
		case err := <-s.coordinator.OnError():
			logl.Error.Printf("coordinator: %s", err)

			// TODO: shut down the system? are there non-fatal coordinator errors?
		case announcedDevice := <-s.coordinator.OnDeviceAnnounce():
			s.registrationQueue <- announcedDevice
		case deviceLeave := <-s.coordinator.OnDeviceLeave():
			ieeeAddress := deviceLeave.ExtAddr

			logl.Info.Printf("Unregistering device: [%s]", ieeeAddress)

			if err := s.unregisterDevice(ieeeAddress); err != nil {
				logl.Error.Printf("unregisterDevice: %s", err.Error())
			}
		case msg := <-s.coordinator.OnDeviceTc():
			logl.Debug.Printf("device online change: %s", msg.SrcIEEEAddr)
		case incomingMessage := <-s.coordinator.OnIncomingMessage():
			if err := s.processIncomingMessage(incomingMessage); err != nil {
				logl.Error.Println(err)
			}
		}
	}
}

func (s *Stack) Channels() *Channels {
	return s.channels
}

func (f *Stack) LocalCommand(dev DeviceAndEndpoint, command cluster.LocalCommand) error {
	clusterId, commandId := command.CommandClusterAndId()

	frm, err := frame.New().
		DisableDefaultResponse(false).
		FrameType(frame.FrameTypeLocal).
		Direction(frame.DirectionClientServer).
		CommandId(commandId).
		Command(command).
		Build()
	if err != nil {
		return err
	}

	response, err := f.coordinator.DataRequest(
		dev.NetworkAddress,
		dev.EndpointId,
		1,
		uint16(clusterId),
		&znp.AfDataRequestOptions{},
		15,
		binstruct.Encode(frm))
	if err != nil {
		return err
	}

	zclIncomingMessage, err := f.zcl.ToZclIncomingMessage(response)
	if err != nil {
		logl.Error.Printf("Unsupported data response message:\n%s\n", spew.Sdump(response))
		return err
	}

	zclCommand := zclIncomingMessage.Data.Command.(*cluster.DefaultResponseCommand)
	if err := zclCommand.Status.Error(); err != nil {
		return fmt.Errorf("unable to run command [%d] on cluster [%d]. Status: %v", commandId, clusterId, err)
	}

	return nil
}

func (s *Stack) Network() *coordinator.Network {
	return s.coordinator.Network()
}

func (s *Stack) processIncomingMessage(incomingMessage *znp.AfIncomingMessage) error {
	zclIncomingMessage, err := s.zcl.ToZclIncomingMessage(incomingMessage)
	if err != nil {
		return fmt.Errorf("Unsupported incoming message: %w: %s", err, spew.Sdump(incomingMessage))
	}

	device, ok := s.db.GetDeviceByNetworkAddress(incomingMessage.SrcAddr)
	if !ok {
		return fmt.Errorf("Received message from unknown device: %s", incomingMessage.SrcAddr)
	}

	select {
	case s.channels.onDeviceIncomingMessage <- &DeviceIncomingMessage{
		Device:          device,
		IncomingMessage: zclIncomingMessage,
	}:
		return nil
	default:
		return errors.New("onDeviceIncomingMessage channel has no capacity. Maybe channel has no subscribers")
	}
}

func (s *Stack) registerDevice(announcedDevice *znp.ZdoEndDeviceAnnceInd) error {
	logl.Info.Printf("Registering device [%s]", announcedDevice.IEEEAddr)

	if device, alreadyExists := s.db.GetDevice(announcedDevice.IEEEAddr); alreadyExists {
		logl.Debug.Printf("device %s already exists in DB. Updating network address", announcedDevice.IEEEAddr)

		// updating NwkAddr because when re-joining, device most likel has changed its network address,
		// (but not its IEEEAddr, e.g. "MAC address")
		device.NetworkAddress = announcedDevice.NwkAddr

		if err := s.db.InsertDevice(device); err != nil {
			return fmt.Errorf("InsertDevice: %w", err)
		}

		select {
		case s.channels.onDeviceBecameAvailable <- device:
			return nil
		default:
			return errors.New("onDeviceBecameAvailable channel has no capacity. Maybe channel has no subscribers")
		}
	}

	device, err := s.interrogateDevice(announcedDevice)
	if err != nil {
		return fmt.Errorf("interrogateDevice: %w", err)
	}

	if err := s.db.InsertDevice(device); err != nil {
		return fmt.Errorf("InsertDevice: %w", err)
	}

	select {
	case s.channels.onDeviceRegistered <- device:
		logl.Info.Printf(
			"Registered new device [%s]. Manufacturer: [%s], Model: [%s], Logical type: [%s]",
			device.IEEEAddress,
			device.Manufacturer,
			device.Model,
			device.LogicalType)

		return nil
	default:
		return errors.New("onDeviceRegistered channel has no capacity. Maybe channel has no subscribers")
	}
}

func (s *Stack) unregisterDevice(ieeeAddress string) error {
	device, found := s.db.GetDevice(ieeeAddress)
	if !found {
		return fmt.Errorf("not found: %s", ieeeAddress)
	}

	if err := s.db.RemoveDevice(ieeeAddress); err != nil {
		return err
	}

	select {
	case s.channels.onDeviceUnregistered <- device:
		logl.Info.Printf(
			"Unregistered device [%s]. Manufacturer: [%s], Model: [%s], Logical type: [%s]",
			ieeeAddress,
			device.Manufacturer,
			device.Model,
			device.LogicalType)

		return nil
	default:
		return errors.New("channel has no capacity. Maybe channel has no subscribers")
	}
}

func castClusterIds(clusterIdsInt []uint16) []cluster.ClusterId {
	clusterIds := []cluster.ClusterId{}
	for _, clusterId := range clusterIdsInt {
		clusterIds = append(clusterIds, cluster.ClusterId(clusterId))
	}

	return clusterIds
}

func openPort(portName string, baudRate int) (port serial.Port, err error) {
	port, err = serial.Open(portName, &serial.Mode{BaudRate: baudRate})
	if err != nil {
		return nil, err
	}

	return port, port.SetRTS(true)
}
