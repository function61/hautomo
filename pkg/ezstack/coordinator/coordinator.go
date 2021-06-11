package coordinator

import (
	"context"
	"encoding/json"
	"fmt"
	"math/bits"
	"reflect"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/function61/gokit/log/logex"
	"github.com/function61/gokit/sync/taskrunner"
	"github.com/function61/hautomo/pkg/ezstack/zcl/frame"
	"github.com/function61/hautomo/pkg/ezstack/zigbee"
	"github.com/function61/hautomo/pkg/ezstack/znp"
	"github.com/tv42/topic"
)

var log = logex.Levels(logex.Prefix("coordinator", logex.StandardLogger()))

// assigns unique numbers like 1, 2, 3, 4, 5... for sent Zigbee commands
var nextTransactionId = frame.MakeDefaultTransactionIdProvider()

const defaultTimeout = 10 * time.Second

type Network struct {
	Address string
}

type MessageChannels struct {
	onError           chan error
	onDeviceAnnounce  chan *znp.ZdoEndDeviceAnnceInd
	onDeviceLeave     chan *znp.ZdoLeaveInd
	onDeviceTc        chan *znp.ZdoTcDevInd
	onIncomingMessage chan *znp.AfIncomingMessage
}

type Coordinator struct {
	config           *Configuration
	networkProcessor *znp.Znp
	messageChannels  *MessageChannels
	network          *Network
	allIncomingMsgs  *topic.Topic // is buffered (capacity ~100)
}

func (c *Coordinator) OnIncomingMessage() chan *znp.AfIncomingMessage {
	return c.messageChannels.onIncomingMessage
}

func (c *Coordinator) OnDeviceTc() chan *znp.ZdoTcDevInd {
	return c.messageChannels.onDeviceTc
}

func (c *Coordinator) OnDeviceLeave() chan *znp.ZdoLeaveInd {
	return c.messageChannels.onDeviceLeave
}

func (c *Coordinator) OnDeviceAnnounce() chan *znp.ZdoEndDeviceAnnceInd {
	return c.messageChannels.onDeviceAnnounce
}

func (c *Coordinator) OnError() chan error {
	return c.messageChannels.onError
}

func (c *Coordinator) Network() *Network {
	return c.network
}

func New(config *Configuration) *Coordinator {
	messageChannels := &MessageChannels{
		onError:           make(chan error, 100),
		onDeviceAnnounce:  make(chan *znp.ZdoEndDeviceAnnceInd, 100),
		onDeviceLeave:     make(chan *znp.ZdoLeaveInd, 100),
		onDeviceTc:        make(chan *znp.ZdoTcDevInd, 100),
		onIncomingMessage: make(chan *znp.AfIncomingMessage, 100),
	}
	return &Coordinator{
		config:          config,
		messageChannels: messageChannels,
		network:         &Network{},
		allIncomingMsgs: topic.New(),
	}
}

func (c *Coordinator) Run(ctx context.Context, joinEnable bool, networkProcessor *znp.Znp, settingsFlash bool) error {
	c.networkProcessor = networkProcessor

	tasks := taskrunner.New(ctx, logex.Discard)

	// this event loop needs to be started early, because the following lines use it, e.g. configureAndReset()
	// calls Reset() which does an async command which blocks on c.allIncomingMsgs.Broadcast recv
	tasks.Start("loop", func(ctx context.Context) error {
		// forward (Errors, AsyncInbound) chans from ZNP to our consumer + do broadcasts
		for {
			select {
			case <-ctx.Done():
				return nil
			case err := <-c.networkProcessor.Errors():
				c.messageChannels.onError <- err
			case incoming := <-c.networkProcessor.AsyncInbound():
				// syncRequestResponse(), syncDataRequestResponse() use this
				c.allIncomingMsgs.Broadcast <- incoming

				switch message := incoming.(type) {
				case *znp.ZdoEndDeviceAnnceInd:
					// "hello, I would like to join the network"
					c.messageChannels.onDeviceAnnounce <- message
				case *znp.ZdoLeaveInd:
					// "hello, I would like to leave the network"
					// some (most?) devices when you initiate pairing, gracefully notify the old network
					// that they're leaving.
					c.messageChannels.onDeviceLeave <- message
				case *znp.ZdoTcDevInd:
					// not exactly sure what this is. I witnessed IKEA bulbs emit this when they already
					// are in the network, and they come back on after being powered off.
					// is this a "I'm back, bitches!" -type of message?
					c.messageChannels.onDeviceTc <- message
				case *znp.AfIncomingMessage:
					// most "application-level" messages, i.e. when a device reports new sensor valies.
					c.messageChannels.onIncomingMessage <- message
				case *znp.SysResetInd, *znp.ZdoStateChangeInd, *znp.ZdoPermitJoinInd, *znp.ZdoMgmtPermitJoinRsp:
					// NO-OP: these are responses to reset, permit join etc. (intercepted by broadcast I guess)
				case *znp.AfDataConfirm:
					// NO-OP: confirmations to app-level requests (intercepted by broadcast I guess)
				case *znp.ZdoSrcRtgInd:
					// NO-OP: at least IKEA bulbs seem to emit these when they are sent commands to.
					//        it contains a relay list.
				case *znp.ZdoNodeDescRsp, *znp.ZdoActiveEpRsp, *znp.ZdoSimpleDescRsp:
					// these are related to when we add a new device
					log.Debug.Print("got probably interview-related messages")
				default:
					// TODO
					log.Error.Printf("unexpected message type: %s", spew.Sdump(incoming))
				}
			}
		}
	})

	firmwareVer, err := c.networkProcessor.SysVersion()
	if err != nil {
		return fmt.Errorf("SysVersion: %w", err)
	}

	log.Info.Printf("starting, firmware v%d.%d.%d (transport v%d)",
		firmwareVer.MajorRel,
		firmwareVer.MinorRel,
		firmwareVer.MaintRel,
		firmwareVer.TransportRev)

	// applies network settings (network address, radio channel, encryption key) etc.
	if err := configureAndReset(c, settingsFlash); err != nil {
		return fmt.Errorf("configureAndReset: %w", err)
	}

	// "enable all subsystems"
	if _, err := c.networkProcessor.UtilCallbackSubCmd(znp.SubsystemIdAllSubsystems, znp.ActionEnable); err != nil {
		return fmt.Errorf("UtilCallbackSubCmd: %w", err)
	}

	// "start zigbee"
	if _, err := c.networkProcessor.SapiZbStartRequest(); err != nil {
		return fmt.Errorf("SapiZbStartRequest: %w", err)
	}

	deviceInfo, err := c.networkProcessor.UtilGetDeviceInfo()
	if err != nil {
		return fmt.Errorf("UtilGetDeviceInfo: %w", err)
	}

	log.Debug.Printf(
		"ZNP DeviceInfo: status=%s IEEEAddr=%s Addr=%s state=%s assocdevices=%v",
		deviceInfo.Status,
		deviceInfo.IEEEAddr,
		deviceInfo.ShortAddr,
		deviceInfo.DeviceState,
		deviceInfo.AssocDevicesList)

	c.network.Address = deviceInfo.ShortAddr // this seems to be 0x0000

	if err := setLed(c.config.Led, c.networkProcessor); err != nil {
		return fmt.Errorf("setLed: %w", err)
	}

	// register endpoints
	// TODO: shimmeringbee works with just zigbee.ProfileHomeAutomation - do we need all these?
	for idx, profileId := range []zigbee.ProfileID{
		zigbee.ProfileHomeAutomation,            // will get assigned endpoint 1 ..
		zigbee.ProfileIndustrialPlantMonitoring, // .. endpoint 2 and so on ..
		zigbee.ProfileCommercialBuildingAutomation,
		zigbee.ProfileTelecomApplications,
		zigbee.ProfilePersonalHomeAndHospitalCare,
		zigbee.ProfileAdvancedMeteringInitiative,
	} {
		endpoint := zigbee.EndpointId(idx + 1) // 1,2,3,...

		if _, err := c.networkProcessor.AfRegister(endpoint, uint16(profileId), 0x0005, 0x1, znp.LatencyNoLatency, []uint16{}, []uint16{}); err != nil {
			return fmt.Errorf("AfRegister(%d, %d): %w", endpoint, profileId, err)
		}
	}

	if joinEnable {
		log.Info.Println("WARN: permitting joining")
	}

	if err := setPermitJoiningStatus(joinEnable, c); err != nil {
		return fmt.Errorf("setPermitJoiningStatus: %w", err)
	}

	log.Info.Println("running")

	/* poll for when joining gets disabled, but I didn't find how to get current joining status
	tasks.Start("nwinfodbg",func(ctx context.Context)error{
		ticker:=time.NewTicker(5*time.Second)

		for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:

		}
		}
		return nil
	})
	*/

	return tasks.Wait()
}

func (c *Coordinator) syncRequestResponse(
	sendRequest func() error,
	expectedType reflect.Type,
	timeout time.Duration,
) (interface{}, error) {
	allIncomingMsgs := make(chan interface{}, 100) // 100 b/c if channel becomes full while we race to consume, we won't get messages
	c.allIncomingMsgs.Register(allIncomingMsgs)
	defer c.allIncomingMsgs.Unregister(allIncomingMsgs)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// calling this should yield a message (of *expectedType*) in allIncomingMsgs
	if err := sendRequest(); err != nil {
		return nil, err
	}

	// try to listen for the timeout duration all incoming messages ("broadcast") whether one comes
	// in matching expectedType
	return func() (interface{}, error) {
		for {
			select {
			case msg := <-allIncomingMsgs:
				if reflect.TypeOf(msg) == expectedType {
					return msg, nil
				}
			case <-ctx.Done():
				return nil, fmt.Errorf("timeout. didn't receive response of type: %s", expectedType)
			}
		}
	}()
}

func (c *Coordinator) syncRequestResponseRetryable(call func() error, expectedType reflect.Type, timeout time.Duration, retries int) (interface{}, error) {
	response, err := c.syncRequestResponse(call, expectedType, timeout)
	switch {
	case err != nil && retries > 0:
		log.Error.Printf("%s. Retries left: %d", err, retries)
		return c.syncRequestResponseRetryable(call, expectedType, timeout, retries-1)
	case err != nil && retries == 0:
		log.Error.Printf("failure: %s", err)
		return nil, err
	}
	return response, nil
}

func (c *Coordinator) syncDataRequestResponseRetryable(request func(string, uint8) error, nwkAddress string, transactionId uint8, timeout time.Duration, retries int) (*znp.AfIncomingMessage, error) {
	incomingMessage, err := c.syncDataRequestResponse(request, nwkAddress, transactionId, timeout)
	switch {
	case err != nil && retries > 0:
		log.Error.Printf("%s. Retries left: %d", err, retries)
		return c.syncDataRequestResponseRetryable(request, nwkAddress, transactionId, timeout, retries-1)
	case err != nil && retries == 0:
		log.Error.Printf("failure: %s", err)
		return nil, err
	}
	return incomingMessage, nil
}

func (c *Coordinator) syncDataRequestResponse(
	sendRequest func(string, uint8) error,
	nwkAddress string,
	transactionId uint8,
	timeout time.Duration,
) (*znp.AfIncomingMessage, error) {
	allIncomingMsgs := make(chan interface{}, 100) // 100 b/c if channel becomes full while we race to consume, we won't get messages
	c.allIncomingMsgs.Register(allIncomingMsgs)
	defer c.allIncomingMsgs.Unregister(allIncomingMsgs)

	// this should yield messages for the allIncomingMsgs:
	// 1) we expect AfDataConfirm with matching transactionId
	// 2) then AfIncomingMessage with matching transactionId
	if err := sendRequest(nwkAddress, transactionId); err != nil {
		return nil, fmt.Errorf("unable to send data request: %s", err)
	}

	ctx1, cancel1 := context.WithTimeout(context.Background(), timeout)
	defer cancel1()

	// 1) AfDataConfirm. this could be a response directly generated by the ZNP to say
	//    e.g. StatusNwkNoRoute if the target device is not currently in the network

	if err := func() error {
		for {
			select {
			case msg := <-allIncomingMsgs:
				if dataConfirm, ok := msg.(*znp.AfDataConfirm); ok && dataConfirm.TransID == transactionId {
					if dataConfirm.Status == znp.StatusSuccess {
						return nil
					} else {
						return fmt.Errorf("data confirm: %s", dataConfirm.Status)
					}
				}
			case <-ctx1.Done():
				return fmt.Errorf("timeout. didn't receive confirmation for transaction: %d", transactionId)
			}
		}
	}(); err != nil {
		return nil, err
	}

	// 2) AfIncomingMessage

	ctx2, cancel2 := context.WithTimeout(context.Background(), timeout) // "extend" timeout
	defer cancel2()

	return func() (*znp.AfIncomingMessage, error) {
		for {
			select {
			case msg := <-allIncomingMsgs:
				if msg, ok := msg.(*znp.AfIncomingMessage); ok {
					frm, err := frame.Decode(msg.Data)
					if err != nil {
						return nil, err
					}

					if frm.TransactionSequenceNumber == transactionId && msg.SrcAddr == nwkAddress {
						return msg, nil
					}
				}
			case <-ctx2.Done():
				return nil, fmt.Errorf("timeout. didn't receive response for transaction: %d", transactionId)
			}
		}
	}()
}

func readNetworkConfigFromNVRAM(np *znp.Znp) (*NetworkConfiguration, error) {
	savedNetworkParams, err := (&znp.UtilGetNvInfoRequest{}).Send(np)
	if err != nil {
		return nil, err
	}

	// no direct util function for this?
	savedExtPanID := znp.ZCDNVExtPANID{}
	if err := np.NVRAMRead(&savedExtPanID); err != nil {
		return nil, err
	}

	/*
		// describes which fields are set
		fields := savedNetworkParameters.Status // shorthand

		assertPresent := func(status *znp.Status, fieldName string) error { // helper
			return nil
				// if *status!=1 {
				// 	return fmt.Errorf("not persisted: %s",fieldName)
				// } else {
				// 	return nil
				// }
		}
	*/

	// the radio supports "scanning" multiple channels, but our logic supports only one channel so
	// assert that so we don't head into unknown waters
	scanChannelsCount := bits.OnesCount(uint(savedNetworkParams.ScanChannels))
	if scanChannelsCount != 1 {
		return nil, fmt.Errorf("scanChannelsCount: %d", scanChannelsCount)
	}

	// after reversing, channel 11 is stored as
	// 0b100000000000, so we just count the rightmost zeros
	channel := bits.TrailingZeros32(bits.ReverseBytes32(savedNetworkParams.ScanChannels))

	/* going to other direction would be:

	channelBits := make([]byte, 4)
	binary.LittleEndian.PutUint32(x, 1<<ch)
	*/

	return &NetworkConfiguration{
		IEEEAddress: zigbee.IEEEAddress(savedNetworkParams.IEEEAddr),
		PanId:       savedNetworkParams.PanID,
		ExtPanId:    savedExtPanID.ExtendedPANID,
		Channel:     uint8(channel),
		NetworkKey:  savedNetworkParams.PreConfigKey[:],
	}, nil
}

func configureAndReset(coordinator *Coordinator, settingsFlash bool) error {
	maybeWrapErr := func(prefix string, err error) error {
		if err != nil {
			return fmt.Errorf("%s%w", prefix, err)
		} else {
			return nil
		}
	}

	if err := coordinator.Reset(); err != nil {
		return maybeWrapErr("Reset: ", err)
	}

	np := coordinator.networkProcessor // shorthand

	// TODO: does the ZNP expect local or UTC time?
	now := time.Now()

	if _, err := np.SysSetTime(0, uint8(now.Hour()), uint8(now.Minute()), uint8(now.Second()),
		uint8(now.Month()), uint8(now.Day()), uint16(now.Year())); err != nil {
		return maybeWrapErr("SysSetTime: ", err)
	}

	radioConf, err := readNetworkConfigFromNVRAM(np)
	if err != nil {
		return fmt.Errorf("readNetworkConfigFromNVRAM: %w", err)
	}

	// check if in-device persisted configuration matches what we have. it would not be
	// wise to flash config each time we start, because it wears out the EEPROM
	if coordinator.config.Equal(*radioConf) {
		return nil // happy path
	}

	if !settingsFlash { // dangerous unless explicitly given permission
		radioConfJson, _ := json.MarshalIndent(radioConf, "", "  ")

		return fmt.Errorf("mismatching config. add flag to allow flashing!\nradio config = %s", radioConfJson)
	}

	log.Error.Println("mismatching config - flashing")

	for _, step := range []func() error{
		func() error {
			_, err := np.UtilSetPreCfgKey(coordinator.config.NetworkKey)

			return maybeWrapErr("UtilSetPreCfgKey: ", err)
		},
		func() error {
			// magic const = (&znp.ZCDNVLogicalType{}).ItemID()
			_, err := np.SapiZbWriteConfiguration(0x87, []uint8{uint8(zigbee.LogicalTypeCoordinator)})

			return maybeWrapErr("SapiZbWriteConfiguration: ", err)
		},
		func() error {
			_, err := np.UtilSetPanId(coordinator.config.PanId)

			return maybeWrapErr("UtilSetPanId: ", err)
		},
		func() error {
			err := np.NVRAMWrite(&znp.ZCDNVExtPANID{
				ExtendedPANID: coordinator.config.ExtPanId,
			})

			return maybeWrapErr("ZCDNVExtPANID: ", err)
		},
		func() error {
			// magic const = (&znp.ZCDNVZDODirectCB{}).ItemID()
			_, err := np.SapiZbWriteConfiguration(0x8F, []uint8{1})

			return maybeWrapErr("SapiZbWriteConfiguration: ", err)
		},
		func() error { // "enable security"
			// magic const = (&znp.ZCDNVSecurityMode{}).ItemID()
			_, err := np.SapiZbWriteConfiguration(0x64, []uint8{1})

			return maybeWrapErr("SapiZbWriteConfiguration: ", err)
		},
		func() error {
			_, err := np.SysSetExtAddr(coordinator.config.IEEEAddress)

			return maybeWrapErr("SysSetExtAddr: ", err)
		},
		func() error {
			// TODO: replace this with bit field
			// TODO: apparently we could list multiple here. is it related to sniffing?
			channels := &znp.Channels{}
			switch coordinator.config.Channel {
			case 11:
				channels.Channel11 = 1
			case 12:
				channels.Channel12 = 1
			case 13:
				channels.Channel13 = 1
			case 14:
				channels.Channel14 = 1
			case 15:
				channels.Channel15 = 1
			case 16:
				channels.Channel16 = 1
			case 17:
				channels.Channel17 = 1
			case 18:
				channels.Channel18 = 1
			case 19:
				channels.Channel19 = 1
			case 20:
				channels.Channel20 = 1
			case 21:
				channels.Channel21 = 1
			case 22:
				channels.Channel22 = 1
			case 23:
				channels.Channel23 = 1
			case 24:
				channels.Channel24 = 1
			case 25:
				channels.Channel25 = 1
			case 26:
				channels.Channel26 = 1
			default:
				return fmt.Errorf("unsupported channel: %d", coordinator.config.Channel)
			}

			_, err := np.UtilSetChannels(channels)

			return maybeWrapErr("UtilSetChannels: ", err)
		},
		func() error {
			return maybeWrapErr("Reset: ", coordinator.Reset())
		},
	} {
		if err := step(); err != nil {
			return err
		}
	}

	return nil
}

func setPermitJoiningStatus(permitJoin bool, coordinator *Coordinator) error {
	// the naming implies that we can enable joining for temporarily, which we do want.
	// TODO: take advantage of the fact.
	timeout := func() uint8 {
		if permitJoin {
			// return 0xFF // forever?
			go func() {
				<-time.After(120 * time.Second)
				log.Info.Println("join period passed")
			}()

			return 120 // minutes? seconds?
		} else {
			return 0x00
		}
	}()

	if _, err := coordinator.networkProcessor.SapiZbPermitJoiningRequest(coordinator.network.Address, timeout); err != nil {
		return fmt.Errorf("SapiZbPermitJoiningRequest: %w", err)
	}

	return nil
}

func setLed(enabled bool, networkProcessor *znp.Znp) error {
	ledMode := func() znp.Mode {
		if enabled {
			return znp.ModeON
		} else {
			return znp.ModeOFF
		}
	}()

	if _, err := networkProcessor.UtilLedControl(1, ledMode); err != nil {
		return fmt.Errorf("UtilLedControl: %w", err)
	}

	return nil
}
