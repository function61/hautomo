package znp

import (
	"fmt"

	"github.com/function61/hautomo/pkg/ezstack/zigbee"
	"github.com/function61/hautomo/pkg/ezstack/znp/unp"
)

// Zigbee Network Processor commands we can send to the radio

// These are probably documented somewhere on the Texas Instruments site, but their documentation site
// is so user-hostile I could not find them.

type StatusResponse struct {
	Status Status
}

// =======AF=======

type AfRegister struct {
	EndPoint          zigbee.EndpointId
	AppProfID         uint16
	AppDeviceID       uint16
	AddDevVer         uint8
	LatencyReq        Latency
	AppInClusterList  []uint16 `size:"1"`
	AppOutClusterList []uint16 `size:"1"`
}

type AfDataRequestOptions struct {
	WildcardProfileID uint8 `bits:"0b00000010" bitmask:"start" `
	APSAck            uint8 `bits:"0b00010000"`
	DiscoverRoute     uint8 `bits:"0b00100000"`
	APSSecurity       uint8 `bits:"0b01000000"`
	SkipRouting       uint8 `bits:"0b10000000" bitmask:"end" `
}

type AfDataRequest struct {
	DstAddr     string `hex:"2"`
	DstEndpoint zigbee.EndpointId
	SrcEndpoint zigbee.EndpointId
	ClusterID   uint16
	TransID     uint8
	Options     *AfDataRequestOptions
	Radius      uint8
	Data        []uint8 `size:"1"`
}

type AfDataRequestExt struct {
	DstAddrMode AddrMode
	DstAddr     string `hex:"8"`
	DstEndpoint zigbee.EndpointId
	DstPanID    uint16 //PAN - personal area networks
	SrcEndpoint zigbee.EndpointId
	ClusterID   uint16
	TransID     uint8
	Options     *AfDataRequestOptions
	Radius      uint8
	Data        []uint8 `size:"2"`
}

type AfDataRequestSrcRtgOptions struct {
	APSAck      uint8 `bits:"0b00000001" bitmask:"start`
	APSSecurity uint8 `bits:"0b00000100"`
	SkipRouting uint8 `bits:"0b00001000" bitmask:"end" `
}

type AfDataRequestSrcRtg struct {
	DstAddr     string `hex:"2"`
	DstEndpoint zigbee.EndpointId
	SrcEndpoint zigbee.EndpointId
	ClusterID   uint16
	TransID     uint8
	Options     *AfDataRequestSrcRtgOptions
	Radius      uint8
	RelayList   []string `size:"1" hex:"2"`
	Data        []uint8  `size:"1"`
}

type AfInterPanCtlData interface {
	AfInterPanCtlData()
}

type AfInterPanClrData struct{}

func (a *AfInterPanClrData) AfInterPanCtlData() {}

type AfInterPanSetData struct {
	Channel uint8
}

func (a *AfInterPanSetData) AfInterPanCtlData() {}

type AfInterPanRegData struct {
	Endpoint zigbee.EndpointId
}

func (a *AfInterPanRegData) AfInterPanCtlData() {}

type AfInterPanChkData struct {
	PanID    uint16
	Endpoint zigbee.EndpointId
}

func (a *AfInterPanChkData) AfInterPanCtlData() {}

type AfInterPanCtl struct {
	Command InterPanCommand
	Data    AfInterPanCtlData
}

type AfDataRetrieve struct {
	Timestamp uint32
	Index     uint16
	Length    uint8
}

type AfDataRetrieveResponse struct {
	Status StatusResponse
	Data   []uint8 `size:"1"`
}

type AfApsfConfigSet struct {
	Endpoint   zigbee.EndpointId
	FrameDelay uint8
	WindowSize uint8
}

type AfDataConfirm struct {
	Status   Status
	Endpoint zigbee.EndpointId
	TransID  uint8
}

type AfReflectError struct {
	Status      Status
	Endpoint    zigbee.EndpointId
	TransID     uint8
	DstAddrMode AddrMode
	DstAddr     string `hex:"2"`
}

type AfIncomingMessage struct {
	GroupID        uint16
	ClusterID      uint16
	SrcAddr        string `hex:"2"`
	SrcEndpoint    zigbee.EndpointId
	DstEndpoint    zigbee.EndpointId
	WasBroadcast   uint8
	LinkQuality    uint8
	SecurityUse    uint8
	Timestamp      uint32
	TransSeqNumber uint8
	Data           []uint8 `size:"1"`
}

type AfDataStore struct {
	Index uint16
	Data  []uint8 `size:"1"`
}

type AfIncomingMessageExt struct {
	GroupID        uint16
	ClusterID      uint16
	SrcAddrMode    AddrMode
	SrcAddr        string `hex:"8"`
	SrcEndpoint    zigbee.EndpointId
	SrcPanID       uint16
	DstEndpoint    zigbee.EndpointId
	WasBroadcast   uint8
	LinkQuality    uint8
	SecurityUse    uint8
	Timestamp      uint32
	TransSeqNumber uint8
	Data           []uint8 `size:"2"`
}

// =======APP=======

type AppMsg struct {
	AppEndpoint zigbee.EndpointId
	DstAddr     string `hex:"2"`
	DstEndpoint zigbee.EndpointId
	ClusterID   uint16
	Message     []uint8 `size:"1"`
}

type AppUserTest struct {
	SrcEndpoint zigbee.EndpointId
	CommandID   uint16
	Parameter1  uint16
	Parameter2  uint16
}

// =======DEBUG=======

type DebugSetThreshold struct {
	ComponentID uint8
	Threshold   uint8
}

type DebugMsg struct {
	String string `size:"1"`
}

// =======SAPI=======

type EmptyResponse struct{}

type SapiZbPermitJoiningRequest struct {
	Destination string `hex:"2"`
	Timeout     uint8  // seconds?
}

type SapiZbBindDevice struct {
	Create      uint8
	CommandID   uint16
	Destination string `hex:"8"`
}

type SapiZbAllowBind struct {
	Timeout uint8
}

type SapiZbSendDataRequest struct {
	Destination string `hex:"2"`
	CommandID   uint16
	Handle      uint8
	Ack         uint8
	Radius      uint8
	Data        []uint8 `size:"1"`
}

type SapiZbReadConfiguration struct {
	ConfigID uint8
}

type SapiZbReadConfigurationResponse struct {
	Status   Status
	ConfigID uint8
	Value    []uint8 `size:"1"`
}

type SapiZbWriteConfiguration struct {
	ConfigID uint8
	Value    []uint8 `size:"1"`
}

type SapiZbGetDeviceInfo struct {
	Param uint8
}

type SapiZbGetDeviceInfoResponse struct {
	Param uint8
	Value uint16
}

type SapiZbFindDeviceRequest struct {
	SearchKey string `hex:"8"`
}

type SapiZbStartConfirm struct {
	Status Status
}

type SapiZbBindConfirm struct {
	CommandID uint16
	Status    Status
}

type SapiZbAllowBindConfirm struct {
	Source string `hex:"2"`
}

type SapiZbSendDataConfirm struct {
	Handle uint8
	Status Status
}

type SapiZbReceiveDataIndication struct {
	Source    string `hex:"2"`
	CommandID uint16
	Data      []uint8 `size:"1"`
}

type SapiZbFindDeviceConfirm struct {
	SearchType uint8
	Result     string `hex:"2"`
	SearchKey  string `hex:"8"`
}

// =======SYS=======

type SysResetReq struct {
	//This command will reset the device by using a hardware reset (i.e.
	//watchdog reset) if ‘Type’ is zero. Otherwise a soft reset (i.e. a jump to the
	//reset vector) is done. This is especially useful in the CC2531, for
	//instance, so that the USB host does not have to contend with the USB
	//H/W resetting (and thus causing the USB host to re-enumerate the device
	//which can cause an open virtual serial port to hang.)
	ResetType byte
}

//Capabilities represents the interfaces that this device can handle (compiled into the device)
type Capabilities struct {
	Sys   uint16 `bitmask:"start" bits:"0x0001"`
	Mac   uint16 `bits:"0x0002"`
	Nwk   uint16 `bits:"0x0004"`
	Af    uint16 `bits:"0x0008"`
	Zdo   uint16 `bits:"0x0010"`
	Sapi  uint16 `bits:"0x0020"`
	Util  uint16 `bits:"0x0040"`
	Debug uint16 `bits:"0x0080"`
	App   uint16 `bits:"0x0100"`
	Zoad  uint16 `bitmask:"end" bits:"0x1000"`
}

type SysPingResponse struct {
	Capabilities *Capabilities
}

type SysVersionResponse struct {
	TransportRev uint8 //Transport protocol revision
	Product      uint8 //Product Id
	MajorRel     uint8 //Software major release number
	MinorRel     uint8 //Software minor release number
	MaintRel     uint8 //Software maintenance release number
}

type SysSetExtAddr struct {
	ExtAddress string `hex:"8"` //The device’s extended address.
}

type SysGetExtAddrResponse struct {
	ExtAddress string `hex:"8"` //The device’s extended address.
}

type SysRamRead struct {
	Address uint16 //Address of the memory that will be read.
	Len     uint8  //The number of bytes that will be read from the target RAM.
}

type SysRamReadResponse struct {
	Status uint8   //Status is either Success (0) or Failure (1).
	Value  []uint8 `size:"1"` //The value read from the target RAM.
}

type SysRamWrite struct {
	Address uint16  //Address of the memory that will be written.
	Value   []uint8 `size:"1"` //The value written to the target RAM.
}

// read a single memory item (identified by ID) from the target non-volatile memory
type SysOsalNvRead struct {
	ID     NVRAMItemId
	Offset uint8 // usually zero
}

func (req *SysOsalNvRead) Send(z *Znp) (rsp *SysOsalNvReadResponse, err error) {
	if err := z.SendSync(unp.S_SYS, 0x08, req, &rsp); err != nil {
		return nil, wrapIfError("SysOsalNvRead", err)
	}

	return rsp, wrapIfError("SysOsalNvRead", rsp.Status.Error())
}

type SysOsalNvReadResponse struct {
	Status Status
	Value  []uint8 `size:"1"`
}

type SysOsalNvWrite struct {
	ID     NVRAMItemId
	Offset uint8  // usually zero
	Value  []byte `size:"1"`
}

// write to a particular item in non-volatile memory
func (req *SysOsalNvWrite) Send(z *Znp) (rsp *StatusResponse, err error) {
	err = z.SendSync(unp.S_SYS, 0x09, req, &rsp)
	return
}

type SysOsalNvItemInit struct {
	ID       uint16
	ItemLen  uint16
	InitData []uint8 `size:"1"`
}

type SysOsalNvDelete struct {
	ID      uint16
	ItemLen uint16
}

type SysOsalNvLength struct {
	ID uint16
}

type SysOsalNvLengthResponse struct {
	Length uint16
}

type SysOsalStartTimer struct {
	ID      uint8
	Timeout uint16
}

type SysOsalStopTimer struct {
	ID uint8
}

type SysRandomResponse struct {
	Value uint16
}

type SysAdcRead struct {
	Channel    Channel
	Resolution Resolution
}

type SysAdcReadResponse struct {
	Value uint16
}

type SysGpio struct {
	Operation Operation
	Value     uint8
}

type SysGpioResponse struct {
	Value uint8
}

type SysTime struct {
	UTCTime uint32
	Hour    uint8
	Minute  uint8
	Second  uint8
	Month   uint8
	Day     uint8
	Year    uint16
}

type SysSetTxPower struct {
	TXPower uint8
}

type SysSetTxPowerResponse struct {
	TXPower uint8
}

type SysZDiagsClearStats struct {
	ClearNV uint8
}

type SysZDiagsClearStatsResponse struct {
	SysClock uint32
}

type SysZDiagsGetStats struct {
	AttributeID uint16
}

type SysZDiagsGetStatsResponse struct {
	AttributeValue uint32
}

type SysZDiagsSaveStatsToNvResponse struct {
	SysClock uint32
}

type SysNvCreate struct {
	SysID  uint8
	ItemID uint16
	SubID  uint16
	Length uint32
}

type SysNvDelete struct {
	SysID  uint8
	ItemID uint16
	SubID  uint16
}

type SysNvLength struct {
	SysID  uint8
	ItemID uint16
	SubID  uint16
}

type SysNvLengthResponse struct {
	Length uint8
}

type SysNvRead struct {
	SysID  uint8
	ItemID uint16
	SubID  uint16
	Offset uint16
	Length uint8
}

type SysNvReadResponse struct {
	Status Status
	Value  []uint8 `size:"1"`
}

type SysNvWrite struct {
	SysID  uint8
	ItemID uint16
	SubID  uint16
	Offset uint16
	Value  []uint8 `size:"1"`
}

type SysNvUpdate struct {
	SysID  uint8
	ItemID uint16
	SubID  uint16
	Value  []uint8 `size:"1"`
}

type SysNvCompact struct {
	Threshold uint16
}

type SysNvReadExt struct {
	ID     uint16
	Offset uint16
}

type SysNvWriteExt struct {
	ID     uint16
	Offset uint16
	Value  []uint8 `size:"1"`
}

type SysResetInd struct {
	Reason       Reason
	TransportRev uint8
	Product      uint8
	MinorRel     uint8
	HwRev        uint8
}

type SysOsalTimerExpired struct {
	ID uint8
}

// =======UTIL=======

type DeviceType struct {
	Coordinator uint8 `bits:"0x01" bitmask:"start"`
	Router      uint8 `bits:"0x02"`
	EndDevice   uint8 `bits:"0x04" bitmask:"end"`
}

type UtilGetDeviceInfoResponse struct {
	Status           Status
	IEEEAddr         string `hex:"8"`
	ShortAddr        string `hex:"2"`
	DeviceType       *DeviceType
	DeviceState      DeviceState
	AssocDevicesList []string `size:"1" hex:"2"`
}

type NvInfoStatus struct {
	IEEEAddress   Status `bits:"0b00000001" bitmask:"start"`
	ScanChannels  Status `bits:"0b00000010"`
	PanID         Status `bits:"0b00000100"`
	SecurityLevel Status `bits:"0b00001000"`
	PreConfigKey  Status `bits:"0b00010000" bitmask:"end"`
}

// read a block of parameters from non-volatile storage of the radio
type UtilGetNvInfoRequest struct{}

func (req *UtilGetNvInfoRequest) Send(z *Znp) (rsp *UtilGetNvInfoResponse, err error) {
	if err := z.SendSync(unp.S_UTIL, 0x01, nil, &rsp); err != nil {
		return nil, wrapIfError("UtilGetNvInfoRequest", err)
	}

	return rsp, nil
}

type UtilGetNvInfoResponse struct {
	Status        *NvInfoStatus
	IEEEAddr      string `hex:"8"` // starts with "0x" (field tag makes it)
	ScanChannels  uint32 // TODO: make have own type
	PanID         zigbee.PANID
	SecurityLevel uint8
	PreConfigKey  zigbee.NetworkKey
}

type UtilSetPanId struct {
	PanID zigbee.PANID
}

type UtilSetChannels struct {
	Channels *Channels
}

type UtilSetSecLevel struct {
	SecLevel uint8
}

type UtilSetPreCfgKey struct {
	PreCfgKey zigbee.NetworkKey
}

type UtilCallbackSubCmd struct {
	SubsystemID SubsystemId
	Action      Action
}

type Keys struct {
	Key1 uint8 `bits:"0x01" bitmask:"start"`
	Key2 uint8 `bits:"0x02"`
	Key3 uint8 `bits:"0x04"`
	Key4 uint8 `bits:"0x08"`
	Key5 uint8 `bits:"0x10"`
	Key6 uint8 `bits:"0x20"`
	Key7 uint8 `bits:"0x40"`
	Key8 uint8 `bits:"0x80" bitmask:"end"`
}

type UtilKeyEvent struct {
	Keys  *Keys
	Shift Shift
}

type UtilTimeAliveResponse struct {
	Seconds uint32
}

type UtilLedControl struct {
	LedID uint8 // 1 is the default LED on CC2531
	Mode  Mode
}

type UtilLoopback struct {
	Data []uint8
}

type UtilDataReq struct {
	SecurityUse uint8
}

type UtilSrcMatchAddEntry struct {
	AddrMode AddrMode
	Address  string `hex:"8"`
	PanID    uint16
}

type UtilSrcMatchDelEntry struct {
	AddrMode AddrMode
	Address  string `hex:"8"`
	PanID    uint16
}

type UtilSrcMatchCheckSrcAddr struct {
	AddrMode AddrMode
	Address  string `hex:"8"`
	PanID    uint16
}

type UtilSrcMatchAckAllPending struct {
	Option Action
}

type UtilSrcMatchCheckAllPendingResponse struct {
	Status Status
	Value  uint8
}

type UtilAddrMgrExtAddrLookup struct {
	ExtAddr string `hex:"8"`
}

type UtilAddrMgrExtAddrLookupResponse struct {
	NwkAddr string `hex:"2"`
}

type UtilAddrMgrAddrLookup struct {
	NwkAddr string `hex:"2"`
}

type UtilAddrMgrAddrLookupResponse struct {
	ExtAddr string `hex:"8"`
}

type UtilApsmeLinkKeyDataGet struct {
	ExtAddr string `hex:"8"`
}

type UtilApsmeLinkKeyDataGetResponse struct {
	Status    Status
	SecKey    [16]uint8
	TxFrmCntr uint32
	RxFrmCntr uint32
}

type UtilApsmeLinkKeyNvIdGet struct {
	ExtAddr string `hex:"8"`
}

type UtilApsmeLinkKeyNvIdGetResponse struct {
	Status      Status
	LinkKeyNvId uint16
}

type UtilApsmeRequestKeyCmd struct {
	PartnerAddr string `hex:"8"`
}

type UtilAssocCount struct {
	StartRelation Relation
	EndRelation   Relation
}

type UtilAssocCountResponse struct {
	Count uint16
}

type LinkInfo struct {
	TxCounter uint8 // Counter of transmission success/failures
	TxCost    uint8 // Average of sending rssi values if link staus is enabled
	// i.e. NWK_LINK_STATUS_PERIOD is defined as non zero
	RxLqi uint8 // average of received rssi values
	// needs to be converted to link cost (1-7) before used
	InKeySeqNum uint8  // security key sequence number
	InFrmCntr   uint32 // security frame counter..
	TxFailure   uint16 // higher values indicate more failures
}

type AgingEndDevice struct {
	EndDevCfg     uint8
	DeviceTimeout uint32
}

type Device struct {
	ShortAddr      string `hex:"2"` // Short address of associated device, or invalid 0xfffe
	AddrIdx        uint16 // Index from the address manager
	NodeRelation   uint8
	DevStatus      uint8 // bitmap of various status values
	AssocCnt       uint8
	Age            uint8
	LinkInfo       *LinkInfo
	EndDev         *AgingEndDevice
	TimeoutCounter uint32
	KeepaliveRcv   uint8
}

type UtilAssocFindDevice struct {
	Number uint8
}

type UtilAssocFindDeviceResponse struct {
	Device *Device
}

type UtilAssocGetWithAddr struct {
	ExtAddr string `hex:"8"`
	NwkAddr string `hex:"2"`
}

type UtilAssocGetWithAddrResponse struct {
	Device *Device
}

type UtilBindAddEntry struct {
	AddrMode    AddrMode
	DstAddr     string `hex:"8"`
	DstEndpoint zigbee.EndpointId
	ClusterIDs  []uint16 `size:"1"`
}

type BindEntry struct {
	SrcEP         uint8
	DstGroupMode  uint8
	DstIdx        uint16
	DstEP         uint8
	ClusterIDList []uint16 `size:"1"`
}

type UtilBindAddEntryResponse struct {
	BindEntry *BindEntry
}

type UtilZclKeyEstInitEst struct {
	TaskID   uint8
	SeqNum   uint8
	EndPoint zigbee.EndpointId
	AddrMode AddrMode
	Addr     string `hex:"8"`
}

type UtilZclKeyEstSign struct {
	Input []uint8 `size:"1"`
}

type UtilZclKeyEstSignResponse struct {
	Status Status
	Key    [42]uint8
}

type UtilSrngGenResponse struct {
	SecureRandomNumbers [100]uint8
}

type UtilSyncReq struct{}

type UtilZclKeyEstablishInd struct {
	TaskId   uint8
	Event    uint8
	Status   uint8
	WaitTime uint8
	Suite    uint16
}

// =======ZDO=======

type ZdoNwkAddrReq struct {
	IEEEAddress string `hex:"8"`
	ReqType     ReqType
	StartIndex  uint8
}

type ZdoIeeeAddrReq struct {
	ShortAddr  string `hex:"2"`
	ReqType    ReqType
	StartIndex uint8
}

type ZdoNodeDescReq struct {
	DstAddr           string `hex:"2"`
	NWKAddrOfInterest string `hex:"2"`
}

type ZdoPowerDescReq struct {
	DstAddr           string `hex:"2"`
	NWKAddrOfInterest string `hex:"2"`
}

type ZdoUserDescReq struct {
	DstAddr           string `hex:"2"`
	NWKAddrOfInterest string `hex:"2"`
}

type ZdoComplexDescReq struct {
	DstAddr           string `hex:"2"`
	NWKAddrOfInterest string `hex:"2"`
}

type ZdoMatchDescReq struct {
	DstAddr           string `hex:"2"`
	NWKAddrOfInterest string `hex:"2"`
	ProfileID         uint16
	InClusterList     []uint16 `size:"1"`
	OutClusterList    []uint16 `size:"1"`
}

type ZdoSimpleDescReq struct {
	DstAddr           string `hex:"2"`
	NWKAddrOfInterest string `hex:"2"`
	Endpoint          zigbee.EndpointId
}

type ZdoActiveEpReq struct {
	DstAddr           string `hex:"2"`
	NWKAddrOfInterest string `hex:"2"`
}

type CapInfo struct {
	AlternatePANCoordinator uint8 `bits:"0b00000001" bitmask:"start"`
	Router                  uint8 `bits:"0b00000010"`
	MainPowered             uint8 `bits:"0b00000100"`
	ReceiverOnWhenIdle      uint8 `bits:"0b00001000"`
	Reserved1               uint8 `bits:"0b00010000"`
	Reserved2               uint8 `bits:"0b00100000"`
	Security                uint8 `bits:"0b01000000"`
	AllocAddr               uint8 `bits:"0b10000000" bitmask:"end"`
}

type ZdoEndDeviceAnnce struct {
	NwkAddr      string `hex:"2"`
	IEEEAddr     string `hex:"8"`
	Capabilities *CapInfo
}

type ZdoUserDescSet struct {
	DstAddr           string `hex:"2"`
	NWKAddrOfInterest string `hex:"2"`
	UserDescriptor    string `size:"1"`
}

type ServerMask struct {
	PrimTrustCenter uint16 `bits:"0x01" bitmask:"start"`
	BkupTrustCenter uint16 `bits:"0x02"`
	PrimBindTable   uint16 `bits:"0x04"`
	BkupBindTable   uint16 `bits:"0x08"`
	PrimDiscTable   uint16 `bits:"0x10"`
	BkupDiscTable   uint16 `bits:"0x20"`
	NetworkManager  uint16 `bits:"0x40" bitmask:"end"`
}

type ZdoServerDiscReq struct {
	ServerMask *ServerMask
}

type ZdoEndDeviceBindReq struct {
	DstAddr              string `hex:"2"`
	LocalCoordinatorAddr string `hex:"2"`
	IEEEAddr             string `hex:"8"`
	Endpoint             zigbee.EndpointId
	ProfileID            uint16
	InClusterList        []uint16 `size:"1"`
	OutClusterList       []uint16 `size:"1"`
}

type ZdoBindUnbindReq struct {
	DstAddr     string `hex:"2"`
	SrcAddress  string `hex:"8"`
	SrcEndpoint zigbee.EndpointId
	ClusterID   uint16
	DstAddrMode AddrMode
	DstAddress  string `hex:"8"` // TODO: this seems to imply we must use DstAddrMode=AddrModeAddr64Bit?
	DstEndpoint zigbee.EndpointId
}

type Channels struct {
	Channel11 uint32 `bits:"0x00000800" bitmask:"start"`
	Channel12 uint32 `bits:"0x00001000"`
	Channel13 uint32 `bits:"0x00002000"`
	Channel14 uint32 `bits:"0x00004000"`
	Channel15 uint32 `bits:"0x00008000"`
	Channel16 uint32 `bits:"0x00010000"`
	Channel17 uint32 `bits:"0x00020000"`
	Channel18 uint32 `bits:"0x00040000"`
	Channel19 uint32 `bits:"0x00080000"`
	Channel20 uint32 `bits:"0x00100000"`
	Channel21 uint32 `bits:"0x00200000"`
	Channel22 uint32 `bits:"0x00400000"`
	Channel23 uint32 `bits:"0x00800000"`
	Channel24 uint32 `bits:"0x01000000"`
	Channel25 uint32 `bits:"0x02000000"`
	Channel26 uint32 `bits:"0b04000000" bitmask:"end"`
}

type ZdoMgmtNwkDiskReq struct {
	DstAddr      string `hex:"2"`
	ScanChannels *Channels
	ScanDuration uint8
	StartIndex   uint8
}

type ZdoMgmtLqiReq struct {
	DstAddr    string `hex:"2"`
	StartIndex uint8
}

type ZdoMgmtRtgReq struct {
	DstAddr    string `hex:"2"`
	StartIndex uint8
}

type ZdoMgmtBindReq struct {
	DstAddr    string `hex:"2"`
	StartIndex uint8
}

type RemoveChildrenRejoin struct {
	Rejoin         uint8 `bits:"0b00000001" bitmask:"start"`
	RemoveChildren uint8 `bits:"0b00000010" bitmask:"end"`
}

type ZdoMgmtLeaveReq struct {
	DstAddr              string `hex:"2"`
	DeviceAddr           string `hex:"8"`
	RemoveChildrenRejoin *RemoveChildrenRejoin
}

type ZdoMgmtDirectJoinReq struct {
	DstAddr    string `hex:"2"`
	DeviceAddr string `hex:"8"`
	CapInfo    *CapInfo
}

type ZdoMgmtPermitJoinReq struct {
	AddrMode       AddrMode
	DstAddr        string `hex:"2"`
	Duration       uint8
	TCSignificance uint8
}

type ZdoMgmtNwkUpdateReq struct {
	DstAddr      string `hex:"2"`
	DstAddrMode  AddrMode
	ChannelMask  *Channels
	ScanDuration uint8
}

type ZdoMsgCbRegister struct {
	ClusterID uint16
}

type ZdoMsgCbRemove struct {
	ClusterID uint16
}

type ZdoStartupFromApp struct {
	StartDelay uint16
}

type ZdoStartupFromAppResponse struct {
	Status StartupFromAppStatus
}

type ZdoSetLinkKey struct {
	ShortAddr   string `hex:"2"`
	IEEEAddr    string `hex:"8"`
	LinkKeyData [16]uint8
}

type ZdoRemoveLinkKey struct {
	IEEEAddr string `hex:"8"`
}

type ZdoGetLinkKey struct {
	IEEEAddr string `hex:"8"`
}

type ZdoGetLinkKeyResponse struct {
	Status      Status
	IEEEAddr    string `hex:"8"`
	LinkKeyData [16]uint8
}

type ZdoNwkDiscoveryReq struct {
	ScanChannels *Channels
	ScanDuration uint8
}

type ZdoJoinReq struct {
	LogicalChannel uint8
	PanID          uint16
	ExtendedPanID  uint64 //64-bit extended PAN ID (ver. 1.1 only). If not v1.1 or don't care, use all 0xFF
	ChosenParent   string `hex:"2"`
	ParentDepth    uint8
	StackProfile   uint8
}

type ZdoSetRejoinParameters struct {
	BackoffDuration uint32
	ScanDuration    uint32
}

type ZdoSecAddLinkKey struct {
	ShortAddress    string `hex:"2"`
	ExtendedAddress string `hex:"8"`
	Key             [16]uint8
}

type ZdoSecEntryLookupExt struct {
	ExtendedAddress string `hex:"8"`
	Entry           [5]uint8
}

type ZdoSecEntryLookupExtResponse struct {
	AMI                  uint16
	KeyNVID              uint16
	AuthenticationOption uint8
}

type ZdoSecDeviceRemove struct {
	ExtendedAddress string `hex:"8"`
}

type ZdoExtRouteDisc struct {
	DestinationAddress string `hex:"2"`
	Options            uint8
	Radius             uint8
}

type ZdoExtRouteCheck struct {
	DestinationAddress string `hex:"2"`
	RTStatus           uint8
	Options            uint8
}

type ZdoExtRemoveGroup struct {
	Endpoint zigbee.EndpointId
	GroupID  uint16
}

type ZdoExtRemoveAllGroup struct {
	Endpoint zigbee.EndpointId
}

type ZdoExtFindAllGroupsEndpoint struct {
	Endpoint  zigbee.EndpointId
	GroupList []uint16 `size:"1"`
}

type ZdoExtFindAllGroupsEndpointResponse struct {
	Groups []uint16 `size:"1"`
}

type ZdoExtFindGroup struct {
	Endpoint zigbee.EndpointId
	GroupID  uint16
}

type ZdoExtFindGroupResponse struct {
	Status  Status
	GroupID uint16
	Name    string `size:"1"`
}

type ZdoExtAddGroup struct {
	Endpoint  zigbee.EndpointId
	GroupID   uint16
	GroupName string `size:"1"`
}

type ZdoExtCountAllGroupsResponse struct {
	Count uint8
}

type ZdoExtRxIdle struct {
	SetFlag  uint8
	SetValue uint8
}

type ZdoExtUpdateNwkKey struct {
	DestinationAddress string `hex:"2"`
	KeySeqNum          uint8
	Key                [128]uint8
}

type ZdoExtSwitchNwkKey struct {
	DestinationAddress string `hex:"2"`
	KeySeqNum          uint8
}

type ZdoExtNwkInfoResponse struct {
	ShortAddress          string `hex:"2"`
	PanID                 uint16
	ParentAddress         string `hex:"2"`
	ExtendedPanID         uint64
	ExtendedParentAddress string `hex:"8"`
	Channel               uint16 //uint16 or uint8?????
}

type ZdoExtSeqApsRemoveReq struct {
	NwkAddress      string `hex:"2"`
	ExtendedAddress string `hex:"8"`
	ParentAddress   string `hex:"2"`
}

type ZdoExtSetParams struct {
	UseMulticast uint8
}

type ZdoNwkAddrOfInterestReq struct {
	DestAddr          string `hex:"2"`
	NwkAddrOfInterest string `hex:"2"`
	Cmd               uint8
}

type ZdoNwkAddrRsp struct {
	Status       Status
	IEEEAddr     string `hex:"8"`
	NwkAddr      string `hex:"2"`
	StartIndex   uint8
	AssocDevList []string `size:"1" hex:"2"`
}

type ZdoIEEEAddrRsp struct {
	Status       Status
	IEEEAddr     string `hex:"8"`
	NwkAddr      string `hex:"2"`
	StartIndex   uint8
	AssocDevList []string `size:"1" hex:"2"`
}

type ZdoNodeDescRsp struct {
	SrcAddr                    string `hex:"2"`
	Status                     Status
	NWKAddrOfInterest          string             `hex:"2"`
	LogicalType                zigbee.LogicalType `bits:"0b00000011" bitmask:"start"`
	ComplexDescriptorAvailable uint8              `bits:"0b00001000"`
	UserDescriptorAvailable    uint8              `bits:"0b00010000"  bitmask:"end"`
	APSFlags                   uint8              `bits:"0b00011111" bitmask:"start"`
	FrequencyBand              uint8              `bits:"0b11100000" bitmask:"end"`
	MacCapabilitiesFlags       *CapInfo
	ManufacturerCode           uint16
	MaxBufferSize              uint8
	MaxInTransferSize          uint16
	ServerMask                 *ServerMask
	MaxOutTransferSize         uint16
	DescriptorCapabilities     uint8
}

type ZdoPowerDescRsp struct {
	SrcAddr                 string `hex:"2"`
	Status                  Status
	NWKAddr                 string `hex:"2"`
	CurrentPowerMode        uint8  `bits:"0b00001111" bitmask:"start"`
	AvailablePowerSources   uint8  `bits:"0b11110000"  bitmask:"end"`
	CurrentPowerSource      uint8  `bits:"0b00001111" bitmask:"start"`
	CurrentPowerSourceLevel uint8  `bits:"0b11110000"  bitmask:"end"`
}

type ZdoSimpleDescRsp struct {
	SrcAddr        string `hex:"2"`
	Status         Status
	NWKAddr        string `hex:"2"`
	Len            uint8
	Endpoint       zigbee.EndpointId
	ProfileID      uint16
	DeviceID       uint16
	DeviceVersion  uint8
	InClusterList  []uint16 `size:"1"`
	OutClusterList []uint16 `size:"1"`
}

type ZdoActiveEpRsp struct {
	SrcAddr      string `hex:"2"`
	Status       Status
	NWKAddr      string              `hex:"2"`
	ActiveEPList []zigbee.EndpointId `size:"1"`
}

type ZdoMatchDescRsp struct {
	SrcAddr   string `hex:"2"`
	Status    Status
	NWKAddr   string  `hex:"2"`
	MatchList []uint8 `size:"1"`
}

type ZdoComplexDescRsp struct {
	SrcAddr           string `hex:"2"`
	Status            Status
	NWKAddr           string `hex:"2"`
	ComplexDescriptor string `size:"1"`
}

type ZdoUserDescRsp struct {
	SrcAddr        string `hex:"2"`
	Status         Status
	NWKAddr        string `hex:"2"`
	UserDescriptor string `size:"1"`
}

type ZdoUserDescConf struct {
	SrcAddr string `hex:"2"`
	Status  Status
	NWKAddr string `hex:"2"`
}

type ZdoServerDiscRsp struct {
	SrcAddr    string `hex:"2"`
	Status     Status
	ServerMask *ServerMask
}

type ZdoEndDeviceBindRsp struct {
	SrcAddr string `hex:"2"`
	Status  Status
}

type ZdoBindRsp struct {
	SrcAddr string `hex:"2"`
	Status  Status
}

type ZdoUnbindRsp struct {
	SrcAddr string `hex:"2"`
	Status  Status
}

type Network struct {
	PanID           uint16 `bound:"8"`
	LogicalChannel  uint8
	StackProfile    uint8 `bits:"0b00001111" bitmask:"start"`
	ZigbeeVersion   uint8 `bits:"0b11110000" bitmask:"end"`
	BeaconOrder     uint8 `bits:"0b00001111" bitmask:"start"`
	SuperFrameOrder uint8 `bits:"0b11110000" bitmask:"end"`
	PermitJoin      uint8
}

type ZdoMgmtNwkDiscRsp struct {
	SrcAddr      string `hex:"2"`
	Status       Status
	NetworkCount uint8
	StartIndex   uint8
	NetworkList  []*Network `size:"1"`
}

type NeighborLqi struct {
	ExtendedPanID   uint64
	ExtendedAddress string        `hex:"8"`
	NetworkAddress  string        `hex:"2"`
	DeviceType      LqiDeviceType `bits:"0b00000011" bitmask:"start"`
	RxOnWhenIdle    uint8         `bits:"0b00001100"`
	Relationship    uint8         `bits:"0b00110000" bitmask:"end"`
	PermitJoining   uint8
	Depth           uint8
	LQI             uint8
}

type ZdoMgmtLqiRsp struct {
	SrcAddr              string `hex:"2"`
	Status               Status
	NeighborTableEntries uint8
	StartIndex           uint8
	NeighborLqiList      []*NeighborLqi `size:"1"`
}

type Route struct {
	DestinationAddress string `hex:"2"`
	Status             RouteStatus
	NextHop            string `hex:"2"`
}

type ZdoMgmtRtgRsp struct {
	SrcAddr             string `hex:"2"`
	Status              Status
	RoutingTableEntries uint8
	StartIndex          uint8
	RoutingTable        []*Route `size:"1"`
}

type Addr struct {
	AddrMode     AddrMode
	ShortAddr    string            `hex:"2" cond:"uint:AddrMode!=3"`
	ExtendedAddr string            `hex:"8" cond:"uint:AddrMode==3"`
	DstEndpoint  zigbee.EndpointId `cond:"uint:AddrMode==3"`
}

type Binding struct {
	SrcAddr     string `hex:"8"`
	SrcEndpoint zigbee.EndpointId
	ClusterID   uint16
	DstAddr     *Addr
}

type ZdoMgmtBindRsp struct {
	SrcAddr          string `hex:"2"`
	Status           Status
	BindTableEntries uint8
	StartIndex       uint8
	BindTable        []*Binding `size:"1"`
}

type ZdoMgmtLeaveRsp struct {
	SrcAddr string `hex:"2"`
	Status  Status
}

type ZdoMgmtDirectJoinRsp struct {
	SrcAddr string `hex:"2"`
	Status  Status
}

type ZdoMgmtPermitJoinRsp struct {
	SrcAddr string `hex:"2"`
	Status  Status
}

type ZdoStateChangeInd struct {
	State DeviceState
}

type ZdoEndDeviceAnnceInd struct {
	SrcAddr      string `hex:"2"`
	NwkAddr      string `hex:"2"`
	IEEEAddr     string `hex:"8"`
	Capabilities *CapInfo
}

type ZdoMatchDescRpsSent struct {
	NwkAddr        string   `hex:"2"`
	InClusterList  []uint16 `size:"1"`
	OutClusterList []uint16 `size:"1"`
}

type ZdoStatusErrorRsp struct {
	SrcAddr string `hex:"2"`
	Status  Status
}

type ZdoSrcRtgInd struct {
	DstAddr   string   `hex:"2"`
	RelayList []string `size:"1" hex:"2"`
}

type Beacon struct {
	SrcAddr         string `hex:"2"`
	PanID           uint16
	LogicalChannel  uint8
	PermitJoining   uint8
	RouterCapacity  uint8
	DeviceCapacity  uint8
	ProtocolVersion uint8
	StackProfile    uint8
	LQI             uint8
	Depth           uint8
	UpdateID        uint8
	ExtendedPanID   uint64
}

type ZdoBeaconNotifyInd struct {
	BeaconList []*Beacon `size:"1"`
}

type ZdoJoinCnf struct {
	Status        Status
	DeviceAddress string `hex:"2"`
	ParentAddress string `hex:"2"`
}

type ZdoNwkDiscoveryCnf struct {
	Status Status
}

type ZdoLeaveInd struct {
	SrcAddr string `hex:"2"`
	ExtAddr string `hex:"8"`
	Request uint8
	Remove  uint8
	Rejoin  uint8
}

type ZdoMsgCbIncoming struct {
	SrcAddr      string `hex:"2"`
	WasBroadcast uint8
	ClusterID    uint16
	SecurityUse  uint8
	SeqNum       uint8
	MacDstAddr   string `hex:"2"`
	Data         []uint8
}

type ZdoTcDevInd struct {
	SrcNwkAddr    string `hex:"2"`
	SrcIEEEAddr   string `hex:"8"`
	ParentNwkAddr string `hex:"2"`
}

type ZdoPermitJoinInd struct {
	PermitJoinDuration uint8
}

// =======APP_CNF=======

type AppCnfSetNwkFrameCounter struct {
	FrameCounterValue uint8
}

type AppCnfSetDefaultEndDeviceTimeout struct {
	Timeout Timeout
}

type AppCnfSetEndDeviceTimeout struct {
	Timeout Timeout
}

type AppCnfSetAllowRejoinTcPolicy struct {
	AllowRejoin uint8
}

type AppCnfBdbStartCommissioning struct {
	CommissioningMode CommissioningMode
}

type AppCnfBdbSetChannel struct {
	IsPrimary uint8
	Channel   *Channels
}

type AppCnfBdbAddInstallCode struct {
	InstallCodeFormat InstallCodeFormat
	IEEEAddr          string `hex:"8"`
	InstallCode       []uint8
}

type AppCnfBdbSetTcRequireKeyExchange struct {
	BdbTrustCenterRequireKeyExchange uint8
}

type AppCnfBdbSetJoinUsesInstallCodeKey struct {
	BdbJoinUsesInstallCodeKey uint8
}

type AppCnfBdbSetActiveDefaultCentralizedKey struct {
	UseGlobal   uint8
	InstallCode [18]uint8
}

type RemainingCommissioningModes struct {
	InitiatorTl    uint8 `bits:"0x01" bitmask:"start"`
	NwkSteering    uint8 `bits:"0x02"`
	NwkFormation   uint8 `bits:"0x04"`
	FindingBinding uint8 `bits:"0x08"`
	Initialization uint8 `bits:"0x10"`
	ParentLost     uint8 `bits:"0x20" bitmask:"end"`
}

type AppCnfBdbCommissioningNotification struct {
	CommissioningStatus         CommissioningStatus
	CommissioningMode           CommissioningMode
	RemainingCommissioningModes *RemainingCommissioningModes
}

// =======GP=======

type TxOptions struct {
	UseGpTxQueue         uint8 `bits:"0b00000001" bitmask:"start"`
	UseCSMAorCA          uint8 `bits:"0b00000010"`
	UseMacAck            uint8 `bits:"0b00000100"`
	GPDFFrameTypeForTx   uint8 `bits:"0b00011000"`
	TxOnMatchingEndpoint uint8 `bits:"0b00100000" bitmask:"end"`
}

type GpDataReq struct {
	Action                 GpAction
	TxOptions              *TxOptions
	ApplicationID          uint8
	SrcID                  uint32
	GPDIEEEAddress         string `hex:"8"`
	Endpoint               zigbee.EndpointId
	GPDCommandID           uint8
	GPDASDU                []uint8 `size:"1"`
	GPEPHandle             uint8
	GPTxQueueEntryLifetime uint32 `bound:"3"`
}

type GpSecRsp struct {
	Status                  GpStatus
	DGPStubHandle           uint8
	ApplicationID           uint8
	SrcID                   uint32
	GPDIEEEAddress          string `hex:"8"`
	Endpoint                zigbee.EndpointId
	GPDFSecurityLevel       uint8
	GPDFKeyType             uint8
	GPDKey                  [16]uint8
	GPDSecurityFrameCounter uint32
}

type GpDataCnf struct {
	Status       Status
	GPMPDUHandle uint8
}

type GpSecReq struct {
	ApplicationID           uint8
	SrcID                   uint32
	GPDIEEEAddress          string `hex:"8"`
	Endpoint                zigbee.EndpointId
	GPDFSecurityLevel       uint8
	GPDFKeyType             uint8
	GPDSecurityFrameCounter uint32
	DGPStubHandle           uint8
}

type GpDataInd struct {
	Status      GpDataIndStatus
	RSSI        uint8
	LinkQuality uint8
	SeqNumber   uint8
	SrcAddrMode AddrMode
	SrcPANId    uint16
	SrcAddress  string `hex:"8"`
	DstAddrMode AddrMode
	DstPANId    uint16
	DstAddress  string  `hex:"8"`
	GPMPDU      []uint8 `size:"1"`
}

func wrapIfError(prefix string, err error) error {
	if err != nil {
		return fmt.Errorf("%s: %w", prefix, err)
	} else {
		return nil
	}
}
