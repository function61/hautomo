package znp

import (
	"fmt"
)

type Latency uint8

const (
	LatencyNoLatency Latency = iota
	LatencyFastBeacons
	LatencySlowBeacons
)

type StartupFromAppStatus uint8

const (
	StartupFromAppStatusRestoredNetworkState StartupFromAppStatus = 0x00
	StartupFromAppStatusNewNetworkState      StartupFromAppStatus = 0x01
	StartupFromAppStatusLeaveAndNotStarted   StartupFromAppStatus = 0x02
)

type Status uint8

// specified in http://software-dl.ti.com/simplelink/esd/plugins/simplelink_zigbee_sdk_plugin/1.60.01.09/exports/docs/zigbee_user_guide/html/zigbee/developing_zigbee_applications/znp_interface/znp_interface.html#return-values
// (our const names have some discrepancies, but non-semantic ones)

const (
	StatusSuccess          Status = 0x00
	StatusFailure          Status = 0x01
	StatusInvalidParameter Status = 0x02

	StatusItemCreatedAndInitialized Status = 0x09
	StatusInitializationFailed      Status = 0x0a
	StatusBadLength                 Status = 0x0c

	// ZStack status values must start at 0x10, after the generic status values (defined in comdef.h)
	StatusMemError        Status = 0x10
	StatusBufferFull      Status = 0x11
	StatusUnsupportedMode Status = 0x12
	StatusMacMemError     Status = 0x13

	StatusSapiInProgress Status = 0x20
	StatusSapiTimeout    Status = 0x21
	StatusSapiInit       Status = 0x22

	StatusNotAuthorized Status = 0x7E

	StatusMalformedCmd    Status = 0x80
	StatusUnsupClusterCmd Status = 0x81

	StatusZdpInvalidEp         Status = 0x82 // Invalid endpoint value
	StatusZdpNotActive         Status = 0x83 // Endpoint not described by a simple desc.
	StatusZdpNotSupported      Status = 0x84 // Optional feature not supported
	StatusZdpTimeout           Status = 0x85 // Operation has timed out
	StatusZdpNoMatch           Status = 0x86 // No match for end device bind
	StatusZdpNoEntry           Status = 0x88 // Unbind request failed, no entry
	StatusZdpNoDescriptor      Status = 0x89 // Child descriptor not available
	StatusZdpInsufficientSpace Status = 0x8a // Insufficient space to support operation
	StatusZdpNotPermitted      Status = 0x8b // Not in proper state to support operation
	StatusZdpTableFull         Status = 0x8c // No table space to support operation
	StatusZdpNotAuthorized     Status = 0x8d // Permissions indicate request not authorized
	StatusZdpBindingTableFull  Status = 0x8e // No binding table space to support operation

	// OTA Status values
	StatusOtaAbort            Status = 0x95
	StatusOtaImageInvalid     Status = 0x96
	StatusOtaWaitForData      Status = 0x97
	StatusOtaNoImageAvailable Status = 0x98
	StatusOtaRequireMoreImage Status = 0x99

	// APS status values
	StatusApsFail              Status = 0xb1
	StatusApsTableFull         Status = 0xb2
	StatusApsIllegalRequest    Status = 0xb3
	StatusApsInvalidBinding    Status = 0xb4
	StatusApsUnsupportedAttrib Status = 0xb5
	StatusApsNotSupported      Status = 0xb6
	StatusApsNoAck             Status = 0xb7
	StatusApsDuplicateEntry    Status = 0xb8
	StatusApsNoBoundDevice     Status = 0xb9
	StatusApsNotAllowed        Status = 0xba
	StatusApsNotAuthenticated  Status = 0xbb

	// Security status values
	StatusSecNoKey       Status = 0xa1
	StatusSecOldFrmCount Status = 0xa2
	StatusSecMaxFrmCount Status = 0xa3
	StatusSecCcmFail     Status = 0xa4
	StatusSecFailure     Status = 0xad

	// NWK status values
	StatusNwkInvalidParam         Status = 0xc1
	StatusNwkInvalidRequest       Status = 0xc2
	StatusNwkNotPermitted         Status = 0xc3
	StatusNwkStartupFailure       Status = 0xc4
	StatusNwkAlreadyPresent       Status = 0xc5
	StatusNwkSyncFailure          Status = 0xc6
	StatusNwkTableFull            Status = 0xc7
	StatusNwkUnknownDevice        Status = 0xc8
	StatusNwkUnsupportedAttribute Status = 0xc9
	StatusNwkNoNetworks           Status = 0xca
	StatusNwkLeaveUnconfirmed     Status = 0xcb
	StatusNwkNoAck                Status = 0xcc // not in spec
	StatusNwkNoRoute              Status = 0xcd

	// MAC status values
	// ZMacSuccess              Status = 0x00
	StatusMacBeaconLoss           Status = 0xe0
	StatusMacChannelAccessFailure Status = 0xe1
	StatusMacDenied               Status = 0xe2
	StatusMacDisableTrxFailure    Status = 0xe3
	StatusMacFailedSecurityCheck  Status = 0xe4
	StatusMacFrameTooLong         Status = 0xe5
	StatusMacInvalidGTS           Status = 0xe6
	StatusMacInvalidHandle        Status = 0xe7
	StatusMacInvalidParameter     Status = 0xe8
	StatusMacNoACK                Status = 0xe9
	StatusMacNoBeacon             Status = 0xea
	StatusMacNoData               Status = 0xeb
	StatusMacNoShortAddr          Status = 0xec
	StatusMacOutOfCap             Status = 0xed
	StatusMacPANIDConflict        Status = 0xee
	StatusMacRealignment          Status = 0xef
	StatusMacTransactionExpired   Status = 0xf0
	StatusMacTransactionOverFlow  Status = 0xf1
	StatusMacTxActive             Status = 0xf2
	StatusMacUnAvailableKey       Status = 0xf3
	StatusMacUnsupportedAttribute Status = 0xf4
	StatusMacUnsupported          Status = 0xf5
	StatusMacSrcMatchInvalidIndex Status = 0xff
)

// converts Status to Go error
func (s Status) Error() error {
	if s != StatusSuccess {
		return fmt.Errorf("ZNP status: %s", s.String())
	}

	return nil
}

type AddrMode uint8

const (
	AddrModeAddrNotPresent AddrMode = iota
	AddrModeAddrGroup
	AddrModeAddr16Bit
	AddrModeAddr64Bit
	AddrModeAddrBroadcast AddrMode = 15 //or 0xFF??????
)

type InterPanCommand uint8

const (
	InterPanCommandInterPanClr InterPanCommand = iota
	InterPanCommandInterPanSet
	InterPanCommandInterPanReg
	InterPanCommandInterPanChk
)

type Channel uint8

const (
	ChannelAIN0 Channel = iota
	ChannelAIN1
	ChannelAIN2
	ChannelAIN3
	ChannelAIN4
	ChannelAIN5
	ChannelAIN6
	ChannelAIN7
	ChannelTemperatureSensor Channel = 0x0E + iota
	ChannelVoltageReading
)

type Resolution uint8

const (
	Resolution8Bit Resolution = iota
	Resolution10Bit
	Resolution12Bit
	Resolution14Bit
)

type Operation uint8

const (
	OperationSetDirection Operation = iota
	OperationSetInputMode
	OperationSet
	OperationClear
	OperationToggle
	OperationRead
)

type Reason uint8

const (
	ReasonPowerUp Reason = iota
	ReasonExternal
	ReasonWatchDog
)

type DeviceState uint8

const (
	DeviceStateInitializedNotStartedAutomatically DeviceState = iota
	DeviceStateInitializedNotConnectedToAnything
	DeviceStateDiscoveringPANsToJoin
	DeviceStateJoiningPAN
	DeviceStateRejoiningPAN
	DeviceStateJoinedButNotAuthenticated
	DeviceStateStartedAsDeviceAfterAuthentication
	DeviceStateDeviceJoinedAuthenticatedAndIsRouter
	DeviceStateStartingAsZigBeeCoordinator
	DeviceStateStartedAsZigBeeCoordinator
	DeviceStateDeviceHasLostInformationAboutItsParent
	DeviceStateDeviceSendingKeepAliveToParent
	DeviceStateDeviceWaitingBeforeRejoin
	DeviceStateReJoiningPANInSecureModeScanningAllChannels
	DeviceStateReJoiningPANInTrustCenterModeScanningCurrentChannel
	DeviceStateReJoiningPANInTrustCenterModeScanningAllChannels
)

type SubsystemId uint16

const (
	SubsystemIdSys           SubsystemId = 0x0100
	SubsystemIdMac           SubsystemId = 0x0200
	SubsystemIdNwk           SubsystemId = 0x0300
	SubsystemIdAf            SubsystemId = 0x0400
	SubsystemIdZdo           SubsystemId = 0x0500
	SubsystemIdSapi          SubsystemId = 0x0600
	SubsystemIdUtil          SubsystemId = 0x0700
	SubsystemIdDebug         SubsystemId = 0x0800
	SubsystemIdApp           SubsystemId = 0x0900
	SubsystemIdAllSubsystems SubsystemId = 0xFFFF
)

type Action uint8

const (
	ActionDisable Action = 0
	ActionEnable  Action = 1
)

type Shift uint8

const (
	ShiftNoShift  Shift = 0
	ShiftYesShift Shift = 1
)

type Mode uint8

const (
	ModeOFF Mode = 0
	ModeON  Mode = 1
)

type Relation uint8

const (
	RelationParent Relation = iota
	RelationChildRfd
	RelationChildRfdRxIdle
	RelationChildFfd
	RelationChildFfdRxIdle
	RelationNeighbor
	RelationOther
)

type ReqType uint8

const (
	ReqTypeSingleDeviceResponse      ReqType = 0x00
	ReqTypeAssociatedDevicesResponse ReqType = 0x01
)

type RouteStatus uint8

const (
	RouteStatusActive            RouteStatus = 0x00
	RouteStatusDiscoveryUnderway RouteStatus = 0x01
	RouteStatusDiscoveryFailed   RouteStatus = 0x02
	RouteStatusInactive          RouteStatus = 0x03
)

type Timeout uint8

const (
	Timeout10Seconds    Timeout = 0x00
	Timeout2Minutes     Timeout = 0x01
	Timeout4Minutes     Timeout = 0x02
	Timeout8Minutes     Timeout = 0x03
	Timeout16Minutes    Timeout = 0x04
	Timeout32Minutes    Timeout = 0x05
	Timeout64Minutes    Timeout = 0x06
	Timeout128Minutes   Timeout = 0x07
	Timeout256Minutes   Timeout = 0x08 //(Default)
	Timeout512Minutes   Timeout = 0x09
	Timeout1024Minutes  Timeout = 0x0A
	Timeout2048Minutes  Timeout = 0x0B
	Timeout4096Minutes  Timeout = 0x0C
	Timeout8192Minutes  Timeout = 0x0D
	Timeout16384Minutes Timeout = 0x0E
)

type InstallCodeFormat uint8

const (
	InstallCodeFormatCodePlusCrc               InstallCodeFormat = 0x00
	InstallCodeFormatKeyDerivedFromInstallCode InstallCodeFormat = 0x01
)

type CommissioningMode uint8

const (
	CommissioningModeInitialization    CommissioningMode = 0x00
	CommissioningModeTouchLink         CommissioningMode = 0x01
	CommissioningModeNetworkSteering   CommissioningMode = 0x02
	CommissioningModeNetworkFormation  CommissioningMode = 0x04
	CommissioningModeFindingAndBinding CommissioningMode = 0x08
)

type CommissioningStatus uint8

const (
	CommissioningStatusSuccess                   CommissioningStatus = 0x00
	CommissioningStatusInProgress                CommissioningStatus = 0x01
	CommissioningStatusNoNetwork                 CommissioningStatus = 0x02
	CommissioningStatusTlTargetFailure           CommissioningStatus = 0x03
	CommissioningStatusTlNotAaCapable            CommissioningStatus = 0x04
	CommissioningStatusTlNoScanResponse          CommissioningStatus = 0x05
	CommissioningStatusTlNotPermitted            CommissioningStatus = 0x06
	CommissioningStatusTclkExFailure             CommissioningStatus = 0x07
	CommissioningStatusFormationFailure          CommissioningStatus = 0x08
	CommissioningStatusFbTargetInProgress        CommissioningStatus = 0x09
	CommissioningStatusFbInitiatorInProgress     CommissioningStatus = 0x0A
	CommissioningStatusFbNoIdentifyQueryResponse CommissioningStatus = 0x0B
	CommissioningStatusFbBindingTableFull        CommissioningStatus = 0x0C
	CommissioningStatusNetwork                   CommissioningStatus = 0x0D
)

type LqiDeviceType uint8

const (
	LqiDeviceTypeCoordinator LqiDeviceType = 0x00
	LqiDeviceTypeRouter      LqiDeviceType = 0x01
	LqiDeviceTypeEndDevice   LqiDeviceType = 0x02
)

type GpAction uint8

const (
	GpActionAddGPDFIntoQueue    GpAction = 0x00
	GpActionRemoveGPDFFromQueue GpAction = 0x01
)

type GpStatus uint8

const (
	GpStatusDropFrame       GpStatus = 0x00
	GpStatusMatch           GpStatus = 0x01
	GpStatusPassUnprocessed GpStatus = 0x02
	GpStatusTxThenDrop      GpStatus = 0x03
	GpStatusError           GpStatus = 0x04
)

type GpDataIndStatus uint8

const (
	GpDataIndStatusSecuritySuccess GpDataIndStatus = 0x00
	GpDataIndStatusNoSecurity      GpDataIndStatus = 0x01
	GpDataIndStatusCounterFailure  GpDataIndStatus = 0x02
	GpDataIndStatusAuthFailure     GpDataIndStatus = 0x03
	GpDataIndStatusUnprocessed     GpDataIndStatus = 0x04
)

const (
	ZbBindingAddr   = "0xFFFE"
	ZbBroadcastAddr = "0xFFFF"
)

const (
	InvalidNodeAddr = "0xFFFE"
)
