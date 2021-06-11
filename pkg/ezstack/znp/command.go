package znp

// TODO: these are essentially 1:1 constructors for structs. refactor AfRegister() to be an
// AfRegister.Send() to have keep benefit of typed responses but reduce duplication of struct fields.
// See *SysOsalNvRead* for example. This file should in time disappear.

import (
	"fmt"

	"github.com/function61/hautomo/pkg/ezstack/zigbee"
	"github.com/function61/hautomo/pkg/ezstack/znp/unp"
)

// =======AF=======

func (znp *Znp) AfRegister(endPoint zigbee.EndpointId, appProfID uint16, appDeviceID uint16, addDevVer uint8,
	latencyReq Latency, appInClusterList []uint16, appOutClusterList []uint16) (rsp *StatusResponse, err error) {
	req := &AfRegister{EndPoint: endPoint, AppProfID: appProfID, AppDeviceID: appDeviceID,
		AddDevVer: addDevVer, LatencyReq: latencyReq, AppInClusterList: appInClusterList, AppOutClusterList: appOutClusterList}
	err = znp.SendSync(unp.S_AF, 0x00, req, &rsp)
	return
}

func (znp *Znp) AfDataRequest(dstAddr string, dstEndpoint zigbee.EndpointId, srcEndpoint zigbee.EndpointId, clusterId uint16,
	transId uint8, options *AfDataRequestOptions, radius uint8, data []uint8) (rsp *StatusResponse, err error) {
	req := &AfDataRequest{DstAddr: dstAddr, DstEndpoint: dstEndpoint, SrcEndpoint: srcEndpoint,
		ClusterID: clusterId, TransID: transId, Options: options, Radius: radius, Data: data}
	err = znp.SendSync(unp.S_AF, 0x01, req, &rsp)
	return
}

func (znp *Znp) AfDataRequestExt(dstAddrMode AddrMode, dstAddr string, dstEndpoint zigbee.EndpointId, dstPanId uint16,
	srcEndpoint zigbee.EndpointId, clusterId uint16, transId uint8, options *AfDataRequestOptions, radius uint8,
	data []uint8) (rsp *StatusResponse, err error) {
	req := &AfDataRequestExt{DstAddrMode: dstAddrMode, DstAddr: dstAddr, DstEndpoint: dstEndpoint, DstPanID: dstPanId, SrcEndpoint: srcEndpoint,
		ClusterID: clusterId, TransID: transId, Options: options, Radius: radius, Data: data}
	err = znp.SendSync(unp.S_AF, 0x02, req, &rsp)
	return
}

func (znp *Znp) AfDataRequestSrcRtg(dstAddr string, dstEndpoint zigbee.EndpointId, srcEndpoint zigbee.EndpointId, clusterId uint16,
	transId uint8, options *AfDataRequestSrcRtgOptions, radius uint8, relayList []string, data []uint8) (rsp *StatusResponse, err error) {
	req := &AfDataRequestSrcRtg{DstAddr: dstAddr, DstEndpoint: dstEndpoint, SrcEndpoint: srcEndpoint,
		ClusterID: clusterId, TransID: transId, Options: options, Radius: radius, RelayList: relayList, Data: data}
	err = znp.SendSync(unp.S_AF, 0x03, req, &rsp)
	return
}

func (znp *Znp) AfInterPanCtl(command InterPanCommand, data AfInterPanCtlData) (rsp *StatusResponse, err error) {
	req := &AfInterPanCtl{Command: command, Data: data}
	err = znp.SendSync(unp.S_AF, 0x10, req, &rsp)
	return
}

func (znp *Znp) AfDataStore(index uint16, data []uint8) (rsp *StatusResponse, err error) {
	req := &AfDataStore{Index: index, Data: data}
	err = znp.SendSync(unp.S_AF, 0x11, req, &rsp)
	return
}

func (znp *Znp) AfDataRetrieve(timestamp uint32, index uint16, length uint8) (rsp *AfDataRetrieveResponse, err error) {
	req := &AfDataRetrieve{Timestamp: timestamp, Index: index, Length: length}
	err = znp.SendSync(unp.S_AF, 0x12, req, &rsp)
	return
}

func (znp *Znp) AfApsfConfigSet(endpoint zigbee.EndpointId, frameDelay uint8, windowSize uint8) (rsp *StatusResponse, err error) {
	req := &AfApsfConfigSet{Endpoint: endpoint, FrameDelay: frameDelay, WindowSize: windowSize}
	err = znp.SendSync(unp.S_AF, 0x13, req, &rsp)
	return
}

// =======APP=======

func (znp *Znp) AppMsg(appEndpoint zigbee.EndpointId, dstAddr string, dstEndpoint zigbee.EndpointId, clusterID uint16,
	message []uint8) (rsp *StatusResponse, err error) {
	req := &AppMsg{AppEndpoint: appEndpoint, DstAddr: dstAddr, DstEndpoint: dstEndpoint,
		ClusterID: clusterID, Message: message}
	err = znp.SendSync(unp.S_APP, 0x00, req, &rsp)
	return
}

func (znp *Znp) AppUserTest(srcEndpoint zigbee.EndpointId, commandId uint16, parameter1 uint16, parameter2 uint16) (rsp *StatusResponse, err error) {
	req := &AppUserTest{SrcEndpoint: srcEndpoint, CommandID: commandId, Parameter1: parameter1, Parameter2: parameter2}
	err = znp.SendSync(unp.S_APP, 0x01, req, &rsp)
	return
}

// =======DEBUG=======

func (znp *Znp) DebugSetThreshold(componentId uint8, threshold uint8) (rsp *StatusResponse, err error) {
	req := &DebugSetThreshold{ComponentID: componentId, Threshold: threshold}
	err = znp.SendSync(unp.S_DBG, 0x00, req, &rsp)
	return
}

func (znp *Znp) DebugMsg(str string) {
	req := &DebugMsg{String: str}
	znp.SendAsync(unp.S_DBG, 0x00, req, nil)
}

// =======MAC======= is not supported on my device

// =======SAPI=======

func (znp *Znp) SapiZbSystemReset() {
	znp.SendAsync(unp.S_SAPI, 0x09, nil, nil)
}

func (znp *Znp) SapiZbStartRequest() (rsp *EmptyResponse, err error) {
	err = znp.SendSync(unp.S_SAPI, 0x00, nil, &rsp)
	return
}

func (znp *Znp) SapiZbPermitJoiningRequest(destination string, timeout uint8) (rsp *StatusResponse, err error) {
	req := &SapiZbPermitJoiningRequest{Destination: destination, Timeout: timeout}
	err = znp.SendSync(unp.S_SAPI, 0x08, req, &rsp)
	return
}

func (znp *Znp) SapiZbBindDevice(create uint8, commandId uint16, destination string) (rsp *EmptyResponse, err error) {
	req := &SapiZbBindDevice{Create: create, CommandID: commandId, Destination: destination}
	err = znp.SendSync(unp.S_SAPI, 0x01, req, &rsp)
	return
}

func (znp *Znp) SapiZbAllowBind(timeout uint8) (rsp *EmptyResponse, err error) {
	req := &SapiZbAllowBind{Timeout: timeout}
	err = znp.SendSync(unp.S_SAPI, 0x02, req, &rsp)
	return
}

func (znp *Znp) SapiZbSendDataRequest(destination string, commandID uint16, handle uint8,
	ack uint8, radius uint8, data []uint8) (rsp *EmptyResponse, err error) {
	req := &SapiZbSendDataRequest{Destination: destination, CommandID: commandID,
		Handle: handle, Ack: ack, Radius: radius, Data: data}
	err = znp.SendSync(unp.S_SAPI, 0x03, req, &rsp)
	return
}

func (znp *Znp) SapiZbReadConfiguration(configID uint8) (rsp *SapiZbReadConfigurationResponse, err error) {
	req := &SapiZbReadConfiguration{ConfigID: configID}
	err = znp.SendSync(unp.S_SAPI, 0x04, req, &rsp)
	return
}

func (znp *Znp) SapiZbWriteConfiguration(configID uint8, value []uint8) (rsp *StatusResponse, err error) {
	req := &SapiZbWriteConfiguration{ConfigID: configID, Value: value}
	err = znp.SendSync(unp.S_SAPI, 0x05, req, &rsp)
	return
}

func (znp *Znp) SapiZbGetDeviceInfo(param uint8) (rsp *SapiZbGetDeviceInfoResponse, err error) {
	req := &SapiZbGetDeviceInfo{Param: param}
	err = znp.SendSync(unp.S_SAPI, 0x06, req, &rsp)
	return
}

func (znp *Znp) SapiZbFindDeviceRequest(searchKey string) (rsp *EmptyResponse, err error) {
	req := &SapiZbFindDeviceRequest{SearchKey: searchKey}
	err = znp.SendSync(unp.S_SAPI, 0x07, req, &rsp)
	return
}

// =======SYS=======

// is sent by the tester to reset the target device
func (znp *Znp) SysResetReq(resetType byte) {
	req := &SysResetReq{resetType}
	znp.SendAsync(unp.S_SYS, 0x00, req, nil)
}

// issues PING requests to verify if a device is active and check the capability of the device.
func (znp *Znp) SysPing() (rsp *SysPingResponse, err error) {
	err = znp.SendSync(unp.S_SYS, 0x01, nil, &rsp)
	return
}

func (znp *Znp) SysVersion() (rsp *SysVersionResponse, err error) {
	err = znp.SendSync(unp.S_SYS, 0x02, nil, &rsp)
	return
}

// set the extended address of the device
func (znp *Znp) SysSetExtAddr(addr zigbee.IEEEAddress) (rsp *StatusResponse, err error) {
	req := &SysSetExtAddr{ExtAddress: addr.HexPrefixedString()}
	err = znp.SendSync(unp.S_SYS, 0x03, req, &rsp)
	return
}

// get the extended address of the device
func (znp *Znp) SysGetExtAddr() (rsp *SysGetExtAddrResponse, err error) {
	err = znp.SendSync(unp.S_SYS, 0x04, nil, &rsp)
	return
}

// read a single memory location in the target RAM. The
// command accepts an address value and returns the memory value present in the target RAM at that address.
func (znp *Znp) SysRamRead(address uint16, len uint8) (rsp *SysRamReadResponse, err error) {
	req := &SysRamRead{Address: address, Len: len}
	err = znp.SendSync(unp.S_SYS, 0x05, req, &rsp)
	return
}

// write to a particular location in the target RAM. The
// command accepts an address location and a memory value. The memory value is written to the
// address location in the target RAM.
func (znp *Znp) SysRamWrite(address uint16, value []uint8) (rsp *StatusResponse, err error) {
	req := &SysRamWrite{Address: address, Value: value}
	err = znp.SendSync(unp.S_SYS, 0x06, req, &rsp)
	return
}

// create and initialize an item in non-volatile memory. The
// NV item will be created if it does not already exist. The data for the new NV item will be left
// uninitialized if the InitLen parameter is zero. When InitLen is non-zero, the data for the NV item
// will be initialized (starting at offset of zero) with the values from InitData. Note that it is not
// necessary to initialize the entire NV item (InitLen < ItemLen). It is also possible to create an NV
// item that is larger than the maximum length InitData – use the SYS_OSAL_NV_WRITE
// command to finish the initialization.
func (znp *Znp) SysOsalNvItemInit(id uint16, itemLen uint16, initData []uint8) (rsp *StatusResponse, err error) {
	req := &SysOsalNvItemInit{ID: id, ItemLen: itemLen, InitData: initData}
	err = znp.SendSync(unp.S_SYS, 0x07, req, &rsp)
	return
}

// delete an item from the non-volatile memory. The ItemLen
// parameter must match the length of the NV item or the command will fail. Use this command with
// caution – deleted items cannot be recovered.
func (znp *Znp) SysOsalNvDelete(id uint16, itemLen uint16) (rsp *StatusResponse, err error) {
	req := &SysOsalNvDelete{ID: id, ItemLen: itemLen}
	err = znp.SendSync(unp.S_SYS, 0x12, req, &rsp)
	return
}

// get the length of an item in non-volatile memory. A
// returned length of zero indicates that the NV item does not exist.
func (znp *Znp) SysOsalNvLength(id uint16) (rsp *SysOsalNvLengthResponse, err error) {
	req := &SysOsalNvLength{ID: id}
	err = znp.SendSync(unp.S_SYS, 0x13, req, &rsp)
	return
}

// start a timer event. The event will expired after the indicated
// amount of time and a notification will be sent back to the tester.
func (znp *Znp) SysOsalStartTimer(id uint8, timeout uint16) (rsp *StatusResponse, err error) {
	req := &SysOsalStartTimer{ID: id, Timeout: timeout}
	err = znp.SendSync(unp.S_SYS, 0x0A, req, &rsp)
	return
}

// stop a timer event.
func (znp *Znp) SysOsalStopTimer(id uint8) (rsp *StatusResponse, err error) {
	req := &SysOsalStopTimer{ID: id}
	err = znp.SendSync(unp.S_SYS, 0x0B, req, &rsp)
	return
}

// get a random 16-bit number.
func (znp *Znp) SysRandom() (rsp *SysRandomResponse, err error) {
	err = znp.SendSync(unp.S_SYS, 0x0C, nil, &rsp)
	return
}

// reads a value from the ADC based on specified channel and resolution.
func (znp *Znp) SysAdcRead(channel Channel, resolution Resolution) (rsp *SysAdcReadResponse, err error) {
	req := &SysAdcRead{Channel: channel, Resolution: resolution}
	err = znp.SendSync(unp.S_SYS, 0x0D, req, &rsp)
	return
}

// control the 4 GPIO pins on the CC2530-ZNP build.
func (znp *Znp) SysGpio(operation Operation, value uint8) (rsp *SysGpioResponse, err error) {
	req := &SysGpio{Operation: operation, Value: value}
	err = znp.SendSync(unp.S_SYS, 0x0E, req, &rsp)
	return
}

// set the target system date and time. The time can be
// specified in “seconds since 00:00:00 on January 1, 2000” or in parsed date/time components
func (znp *Znp) SysSetTime(utcTime uint32, hour uint8, minute uint8, second uint8,
	month uint8, day uint8, year uint16) (rsp *StatusResponse, err error) {
	req := &SysTime{UTCTime: utcTime, Hour: hour, Minute: minute, Second: second, Month: month, Day: day, Year: year}
	err = znp.SendSync(unp.S_SYS, 0x10, req, &rsp)
	return
}

// get the target system date and time. The time is returned in
// seconds since 00:00:00 on January 1, 2000” and parsed date/time components.
func (znp *Znp) SysGetTime() (rsp *SysTime, err error) {
	err = znp.SendSync(unp.S_SYS, 0x11, nil, &rsp)
	return
}

// set the target system radio transmit power. The returned TX
// power is the actual setting applied to the radio – nearest characterized value for the specific radio
func (znp *Znp) SysSetTxPower(txPower uint8) (rsp *SysSetTxPowerResponse, err error) {
	req := &SysSetTxPower{TXPower: txPower}
	err = znp.SendSync(unp.S_SYS, 0x14, req, &rsp)
	return
}

// initialize the statistics table in NV memory.
func (znp *Znp) SysZDiagsInitStats() (rsp *StatusResponse, err error) {
	err = znp.SendSync(unp.S_SYS, 0x17, nil, &rsp)
	return
}

// clear the statistics table. To clear data in NV (including the Boot
// Counter) the clearNV flag shall be set to TRUE.
func (znp *Znp) SysZDiagsClearStats(clearNV uint8) (rsp *SysZDiagsClearStatsResponse, err error) {
	req := &SysZDiagsClearStats{ClearNV: clearNV}
	err = znp.SendSync(unp.S_SYS, 0x18, req, &rsp)
	return
}

// read a specific system (attribute) ID statistics and/or metrics value.
func (znp *Znp) SysZDiagsGetStats(attributeID uint16) (rsp *SysZDiagsGetStatsResponse, err error) {
	req := &SysZDiagsGetStats{AttributeID: attributeID}
	err = znp.SendSync(unp.S_SYS, 0x19, req, &rsp)
	return
}

// restore the statistics table from NV into the RAM table.
func (znp *Znp) SysZDiagsRestoreStatsNv() (rsp *StatusResponse, err error) {
	err = znp.SendSync(unp.S_SYS, 0x1A, nil, &rsp)
	return
}

// save the statistics table from RAM to NV.
func (znp *Znp) SysZDiagsSaveStatsToNv() (rsp *SysZDiagsSaveStatsToNvResponse, err error) {
	err = znp.SendSync(unp.S_SYS, 0x1B, nil, &rsp)
	return
}

// attempt to create an item in non-volatile memory.
func (znp *Znp) SysNvCreate(sysID uint8, itemID uint16, subID uint16, length uint32) (rsp *StatusResponse, err error) {
	req := &SysNvCreate{SysID: sysID, ItemID: itemID, SubID: subID, Length: length}
	err = znp.SendSync(unp.S_SYS, 0x30, req, &rsp)
	return
}

// attempt to delete an item in non-volatile memory.
func (znp *Znp) SysNvDelete(sysID uint8, itemID uint16, subID uint16) (rsp *StatusResponse, err error) {
	req := &SysNvDelete{SysID: sysID, ItemID: itemID, SubID: subID}
	err = znp.SendSync(unp.S_SYS, 0x31, req, &rsp)
	return
}

// get the length of an item in non-volatile memory.
func (znp *Znp) SysNvLength(sysID uint8, itemID uint16, subID uint16) (rsp *SysNvLengthResponse, err error) {
	req := &SysNvLength{SysID: sysID, ItemID: itemID, SubID: subID}
	err = znp.SendSync(unp.S_SYS, 0x32, req, &rsp)
	return
}

// read an item in non-volatile memory
func (znp *Znp) SysNvRead(sysID uint8, itemID uint16, subID uint16, offset uint16, length uint8) (rsp *SysNvReadResponse, err error) {
	req := &SysNvRead{SysID: sysID, ItemID: itemID, SubID: subID, Offset: offset, Length: length}
	err = znp.SendSync(unp.S_SYS, 0x33, req, &rsp)
	return
}

// write an item in non-volatile memory
func (znp *Znp) SysNvWrite(sysID uint8, itemID uint16, subID uint16, offset uint16, value []uint8) (rsp *StatusResponse, err error) {
	req := &SysNvWrite{SysID: sysID, ItemID: itemID, SubID: subID, Offset: offset, Value: value}
	err = znp.SendSync(unp.S_SYS, 0x34, req, &rsp)
	return
}

// update an item in non-volatile memory
func (znp *Znp) SysNvUpdate(sysID uint8, itemID uint16, subID uint16, value []uint8) (rsp *StatusResponse, err error) {
	req := &SysNvUpdate{SysID: sysID, ItemID: itemID, SubID: subID, Value: value}
	err = znp.SendSync(unp.S_SYS, 0x35, req, &rsp)
	return
}

// compact the active page in non-volatile memory
func (znp *Znp) SysNvCompact(threshold uint16) (rsp *StatusResponse, err error) {
	req := &SysNvCompact{Threshold: threshold}
	err = znp.SendSync(unp.S_SYS, 0x36, req, &rsp)
	return
}

// read a single memory item from the target non-volatile
// memory. The command accepts an attribute Id value and data offset and returns the memory value
// present in the target for the specified attribute Id.
func (znp *Znp) SysNvReadExt(id uint16, offset uint16) (rsp *SysNvReadResponse, err error) {
	req := &SysNvReadExt{ID: id, Offset: offset}
	err = znp.SendSync(unp.S_SYS, 0x08, req, &rsp)
	return
}

// write an item in non-volatile memory
func (znp *Znp) SysNvWriteExt(id uint16, offset uint16, value []uint8) (rsp *StatusResponse, err error) {
	req := &SysNvWriteExt{ID: id, Offset: offset, Value: value}
	err = znp.SendSync(unp.S_SYS, 0x09, req, &rsp)
	return
}

// =======UTIL=======

// is sent by the tester to retrieve the device info.
func (znp *Znp) UtilGetDeviceInfo() (rsp *UtilGetDeviceInfoResponse, err error) {
	err = znp.SendSync(unp.S_UTIL, 0x00, nil, &rsp)
	return
}

// stores a PanId value into non-volatile memory to be used the next time the target device resets.
func (znp *Znp) UtilSetPanId(panId zigbee.PANID) (rsp *StatusResponse, err error) {
	req := &UtilSetPanId{PanID: panId}
	err = znp.SendSync(unp.S_UTIL, 0x02, req, &rsp)
	return
}

// store a channel select bit-mask into non-volatile memory to be used the
// next time the target device resets.
func (znp *Znp) UtilSetChannels(channels *Channels) (rsp *StatusResponse, err error) {
	req := &UtilSetChannels{Channels: channels}
	err = znp.SendSync(unp.S_UTIL, 0x03, req, &rsp)
	return
}

// store a security level value into non-volatile memory to be used the next time the target device
// resets.
func (znp *Znp) UtilSetSecLevel(secLevel uint8) (rsp *StatusResponse, err error) {
	req := &UtilSetSecLevel{SecLevel: secLevel}
	err = znp.SendSync(unp.S_UTIL, 0x04, req, &rsp)
	return
}

// store a pre-configured key array into non-volatile memory to be used the
// next time the target device resets.
func (znp *Znp) UtilSetPreCfgKey(key []byte) (rsp *StatusResponse, err error) {
	preCfgKey := [16]byte{}
	if copy(preCfgKey[:], key) != 16 {
		return nil, fmt.Errorf("invalid length for network key: %d", len(key))
	}

	req := &UtilSetPreCfgKey{PreCfgKey: preCfgKey}
	err = znp.SendSync(unp.S_UTIL, 0x05, req, &rsp)
	return
}

// subscribes/unsubscribes to layer callbacks. For particular subsystem callbacks to
// work, the software must be compiled with a special flag that is unique to that subsystem to enable
// the callback mechanism. For example to enable ZDO callbacks, MT_ZDO_CB_FUNC flag must
// be compiled when the software is built. For complete list of callback compile flags, check section
// 1.2 or “Z-Stack Compile Options” document.
func (znp *Znp) UtilCallbackSubCmd(subsystemID SubsystemId, action Action) (rsp *StatusResponse, err error) {
	req := &UtilCallbackSubCmd{SubsystemID: subsystemID, Action: action}
	err = znp.SendSync(unp.S_UTIL, 0x06, req, &rsp)
	return
}

// sends key and shift codes to the application that registered for key events. The keys parameter is a
// bit mask, allowing for multiple keys in a single command. The return status indicates success if
// the command is processed by a registered key handler, not whether the key code was used. Not all
// applications support all key or shift codes but there is no indication when a key code is dropped.
func (znp *Znp) UtilKeyEvent(keys *Keys, shift Shift) (rsp *StatusResponse, err error) {
	req := &UtilKeyEvent{Keys: keys, Shift: shift}
	err = znp.SendSync(unp.S_UTIL, 0x07, req, &rsp)
	return
}

// get the board’s time alive
func (znp *Znp) UtilTimeAlive() (rsp *UtilTimeAliveResponse, err error) {
	err = znp.SendSync(unp.S_UTIL, 0x09, nil, &rsp)
	return
}

// control the LEDs on the board.
func (znp *Znp) UtilLedControl(ledID uint8, mode Mode) (rsp *StatusResponse, err error) {
	req := &UtilLedControl{LedID: ledID, Mode: mode}
	err = znp.SendSync(unp.S_UTIL, 0x0A, req, &rsp)
	return
}

// test data buffer loopback.
func (znp *Znp) UtilLoopback(data []uint8) (rsp *UtilLoopback, err error) {
	req := &UtilLoopback{Data: data}
	err = znp.SendSync(unp.S_UTIL, 0x10, req, &rsp)
	return
}

// effect a MAC MLME Poll Request
func (znp *Znp) UtilDataReq(securityUse uint8) (rsp *StatusResponse, err error) {
	req := &UtilDataReq{SecurityUse: securityUse}
	err = znp.SendSync(unp.S_UTIL, 0x11, req, &rsp)
	return
}

// enable AUTOPEND and source address matching.
func (znp *Znp) UtilSrcMatchEnable() (rsp *StatusResponse, err error) {
	err = znp.SendSync(unp.S_UTIL, 0x20, nil, &rsp)
	return
}

// add a short or extended address to the source address table
func (znp *Znp) UtilSrcMatchAddEntry(addrMode AddrMode, address string, panId uint16) (rsp *StatusResponse, err error) {
	req := &UtilSrcMatchAddEntry{AddrMode: addrMode, Address: address, PanID: panId}
	err = znp.SendSync(unp.S_UTIL, 0x21, req, &rsp)
	return
}

// delete a short or extended address from the source address table.
func (znp *Znp) UtilSrcMatchDelEntry(addrMode AddrMode, address string, panId uint16) (rsp *StatusResponse, err error) {
	req := &UtilSrcMatchDelEntry{AddrMode: addrMode, Address: address, PanID: panId}
	err = znp.SendSync(unp.S_UTIL, 0x22, req, &rsp)
	return
}

// delete a short or extended address from the source address table.
func (znp *Znp) UtilSrcMatchCheckSrcAddr(addrMode AddrMode, address string, panId uint16) (rsp *StatusResponse, err error) {
	req := &UtilSrcMatchCheckSrcAddr{AddrMode: addrMode, Address: address, PanID: panId}
	err = znp.SendSync(unp.S_UTIL, 0x23, req, &rsp)
	return
}

// enable/disable acknowledging all packets with pending bit set.
func (znp *Znp) UtilSrcMatchAckAllPending(option Action) (rsp *StatusResponse, err error) {
	req := &UtilSrcMatchAckAllPending{Option: option}
	err = znp.SendSync(unp.S_UTIL, 0x24, req, &rsp)
	return
}

// check if acknowledging all packets with pending bit set is enabled.
func (znp *Znp) UtilSrcMatchCheckAllPending() (rsp *UtilSrcMatchCheckAllPendingResponse, err error) {
	err = znp.SendSync(unp.S_UTIL, 0x25, nil, &rsp)
	return
}

// is a proxy call to the AddrMgrEntryLookupExt() function.
func (znp *Znp) UtilAddrMgrExtAddrLookup(extAddr string) (rsp *UtilAddrMgrExtAddrLookupResponse, err error) {
	req := &UtilAddrMgrExtAddrLookup{ExtAddr: extAddr}
	err = znp.SendSync(unp.S_UTIL, 0x40, req, &rsp)
	return
}

// is a proxy call to the AddrMgrEntryLookupNwk() function.
func (znp *Znp) UtilAddrMgrAddrLookup(nwkAddr string) (rsp *UtilAddrMgrAddrLookupResponse, err error) {
	req := &UtilAddrMgrAddrLookup{NwkAddr: nwkAddr}
	err = znp.SendSync(unp.S_UTIL, 0x41, req, &rsp)
	return
}

// retrieves APS link key data, Tx and Rx frame counters
func (znp *Znp) UtilApsmeLinkKeyDataGet(extAddr string) (rsp *UtilApsmeLinkKeyDataGetResponse, err error) {
	req := &UtilApsmeLinkKeyDataGet{ExtAddr: extAddr}
	err = znp.SendSync(unp.S_UTIL, 0x44, req, &rsp)
	return
}

// is a proxy call to the APSME_LinkKeyNvIdGet() function.
func (znp *Znp) UtilApsmeLinkKeyNvIdGet(extAddr string) (rsp *UtilApsmeLinkKeyNvIdGetResponse, err error) {
	req := &UtilApsmeLinkKeyNvIdGet{ExtAddr: extAddr}
	err = znp.SendSync(unp.S_UTIL, 0x45, req, &rsp)
	return
}

// send a request key to the Trust Center from an originator device who
// wants to exchange messages with a partner device.
func (znp *Znp) UtilApsmeRequestKeyCmd(partnerAddr string) (rsp *StatusResponse, err error) {
	req := &UtilApsmeRequestKeyCmd{PartnerAddr: partnerAddr}
	err = znp.SendSync(unp.S_UTIL, 0x4B, req, &rsp)
	return
}

// is a proxy call to the AssocCount() function
func (znp *Znp) UtilAssocCount(startRelation Relation, endRelation Relation) (rsp *UtilAssocCountResponse, err error) {
	req := &UtilAssocCount{StartRelation: startRelation, EndRelation: endRelation}
	err = znp.SendSync(unp.S_UTIL, 0x48, req, &rsp)
	return
}

// is a proxy call to the AssocFindDevice() function.
func (znp *Znp) UtilAssocFindDevice(number uint8) (rsp *UtilAssocFindDeviceResponse, err error) {
	req := &UtilAssocFindDevice{Number: number}
	err = znp.SendSync(unp.S_UTIL, 0x49, req, &rsp)
	return
}

// is a proxy call to the AssocGetWithAddress() function.
func (znp *Znp) UtilAssocGetWithAddr(extAddr string, nwkAddr string) (rsp *UtilAssocGetWithAddrResponse, err error) {
	req := &UtilAssocGetWithAddr{ExtAddr: extAddr, NwkAddr: nwkAddr}
	err = znp.SendSync(unp.S_UTIL, 0x4A, req, &rsp)
	return
}

// is a proxy call to the bindAddEntry() function
func (znp *Znp) UtilBindAddEntry(addrMode AddrMode, dstAddr string, dstEndpoint zigbee.EndpointId, clusterIds []uint16) (rsp *UtilBindAddEntryResponse, err error) {
	req := &UtilBindAddEntry{AddrMode: addrMode, DstAddr: dstAddr, DstEndpoint: dstEndpoint, ClusterIDs: clusterIds}
	err = znp.SendSync(unp.S_UTIL, 0x4D, req, &rsp)
	return
}

// is a proxy call to zclGeneral_KeyEstablish_InitiateKeyEstablishment().
func (znp *Znp) UtilZclKeyEstInitEst(taskId uint8, seqNum uint8, endPoint zigbee.EndpointId, addrMode AddrMode, addr string) (rsp *StatusResponse, err error) {
	req := &UtilZclKeyEstInitEst{TaskID: taskId, SeqNum: seqNum, EndPoint: endPoint, AddrMode: addrMode, Addr: addr}
	err = znp.SendSync(unp.S_UTIL, 0x80, req, &rsp)
	return
}

// is a proxy call to zclGeneral_KeyEstablishment_ECDSASign().
func (znp *Znp) UtilZclKeyEstSign(input []uint8) (rsp *UtilZclKeyEstSignResponse, err error) {
	req := &UtilZclKeyEstSign{Input: input}
	err = znp.SendSync(unp.S_UTIL, 0x81, req, &rsp)
	return
}

// generate Secure Random Number. It generates 1,000,000 bits in sets of
// 100 bytes. As in 100 bytes of secure random numbers are generated until 1,000,000 bits are
// generated. 100 bytes are generate
func (znp *Znp) UtilSrngGen() (rsp *UtilSrngGenResponse, err error) {
	err = znp.SendSync(unp.S_UTIL, 0x4C, nil, &rsp)
	return
}

// is an asynchronous request/response handshake.
func (znp *Znp) UtilSyncReq() {
	znp.SendAsync(unp.S_UTIL, 0xE0, nil, nil)
}

// =======ZDO=======

// will request the device to send a “Network Address Request”. This message sends a
// broadcast message looking for a 16 bit address with a known 64 bit IEEE address. You must
// subscribe to “ZDO Network Address Response” to receive the response to this message. Check
// section 3.0.1.7 for more details on callback subscription. The response message listed below only
// indicates whether or not the message was received properly.
func (znp *Znp) ZdoNwkAddrReq(ieeeAddress string, reqType ReqType, startIndex uint8) (rsp *StatusResponse, err error) {
	req := &ZdoNwkAddrReq{IEEEAddress: ieeeAddress, ReqType: reqType, StartIndex: startIndex}
	err = znp.SendSync(unp.S_ZDO, 0x00, req, &rsp)
	return
}

// will request a device’s IEEE 64-bit address. You must subscribe to “ZDO IEEE
// Address Response” to receive the data response to this message. The response message listed
// below only indicates whether or not the message was received properly.
func (znp *Znp) ZdoIeeeAddrReq(shortAddr string, reqType ReqType, startIndex uint8) (rsp *StatusResponse, err error) {
	req := &ZdoIeeeAddrReq{ShortAddr: shortAddr, ReqType: reqType, StartIndex: startIndex}
	err = znp.SendSync(unp.S_ZDO, 0x01, req, &rsp)
	return
}

// inquire about the Node Descriptor information of the destination
// device.
func (znp *Znp) ZdoNodeDescReq(dstAddr string, nwkAddrOfInterest string) (rsp *StatusResponse, err error) {
	req := &ZdoNodeDescReq{DstAddr: dstAddr, NWKAddrOfInterest: nwkAddrOfInterest}
	err = znp.SendSync(unp.S_ZDO, 0x02, req, &rsp)
	return
}

// inquire about the Power Descriptor information of the destination
// device.
func (znp *Znp) ZdoPowerDescReq(dstAddr string, nwkAddrOfInterest string) (rsp *StatusResponse, err error) {
	req := &ZdoPowerDescReq{DstAddr: dstAddr, NWKAddrOfInterest: nwkAddrOfInterest}
	err = znp.SendSync(unp.S_ZDO, 0x03, req, &rsp)
	return
}

// inquire as to the Simple Descriptor of the destination device’s
// Endpoint.
func (znp *Znp) ZdoSimpleDescReq(dstAddr string, nwkAddrOfInterest string, endpoint zigbee.EndpointId) (rsp *StatusResponse, err error) {
	req := &ZdoSimpleDescReq{DstAddr: dstAddr, NWKAddrOfInterest: nwkAddrOfInterest, Endpoint: endpoint}
	err = znp.SendSync(unp.S_ZDO, 0x04, req, &rsp)
	return
}

// request a list of active endpoint from the destination device
func (znp *Znp) ZdoActiveEpReq(dstAddr string, nwkAddrOfInterest string) (rsp *StatusResponse, err error) {
	req := &ZdoActiveEpReq{DstAddr: dstAddr, NWKAddrOfInterest: nwkAddrOfInterest}
	err = znp.SendSync(unp.S_ZDO, 0x05, req, &rsp)
	return
}

// request the device match descriptor
func (znp *Znp) ZdoMatchDescReq(dstAddr string, nwkAddrOfInterest string, profileId uint16,
	inClusterList []uint16, outClusterList []uint16) (rsp *StatusResponse, err error) {
	req := &ZdoMatchDescReq{DstAddr: dstAddr, NWKAddrOfInterest: nwkAddrOfInterest, ProfileID: profileId,
		InClusterList: inClusterList, OutClusterList: outClusterList}
	err = znp.SendSync(unp.S_ZDO, 0x06, req, &rsp)
	return
}

// request for the destination device’s complex descriptor.
func (znp *Znp) ZdoComplexDescReq(dstAddr string, nwkAddrOfInterest string) (rsp *StatusResponse, err error) {
	req := &ZdoComplexDescReq{DstAddr: dstAddr, NWKAddrOfInterest: nwkAddrOfInterest}
	err = znp.SendSync(unp.S_ZDO, 0x07, req, &rsp)
	return
}

// request for the destination device’s user descriptor
func (znp *Znp) ZdoUserDescReq(dstAddr string, nwkAddrOfInterest string) (rsp *StatusResponse, err error) {
	req := &ZdoUserDescReq{DstAddr: dstAddr, NWKAddrOfInterest: nwkAddrOfInterest}
	err = znp.SendSync(unp.S_ZDO, 0x08, req, &rsp)
	return
}

// will cause the device to issue an “End device announce” broadcast packet to the
// network. This is typically used by an end-device to announce itself to the network.
func (znp *Znp) ZdoEndDeviceAnnce(nwkAddr string, ieeeAddr string, capabilities *CapInfo) (rsp *StatusResponse, err error) {
	req := &ZdoEndDeviceAnnce{NwkAddr: nwkAddr, IEEEAddr: ieeeAddr, Capabilities: capabilities}
	err = znp.SendSync(unp.S_ZDO, 0x0A, req, &rsp)
	return
}

// write a User Descriptor value to the targeted device.
func (znp *Znp) ZdoUserDescSet(dstAddr string, nwkAddrOfInterest string, userDescriptor string) (rsp *StatusResponse, err error) {
	req := &ZdoUserDescSet{DstAddr: dstAddr, NWKAddrOfInterest: nwkAddrOfInterest, UserDescriptor: userDescriptor}
	err = znp.SendSync(unp.S_ZDO, 0x0B, req, &rsp)
	return
}

// is used for local device to discover the location of a particular system server or
// servers as indicated by the ServerMask parameter. The destination addressing on this request is
// ‘broadcast to all RxOnWhenIdle devices’.
func (znp *Znp) ZdoServerDiscReq(serverMask *ServerMask) (rsp *StatusResponse, err error) {
	req := &ZdoServerDiscReq{ServerMask: serverMask}
	err = znp.SendSync(unp.S_ZDO, 0x0C, req, &rsp)
	return
}

// request an End Device Bind with the destination device.
func (znp *Znp) ZdoEndDeviceBindReq(dstAddr string, localCoordinatorAddr string, ieeeAddr string, endpoint zigbee.EndpointId,
	profileId uint16, inClusterList []uint16, outClusterList []uint16) (rsp *StatusResponse, err error) {
	req := &ZdoEndDeviceBindReq{DstAddr: dstAddr, LocalCoordinatorAddr: localCoordinatorAddr, IEEEAddr: ieeeAddr,
		Endpoint: endpoint, ProfileID: profileId, InClusterList: inClusterList, OutClusterList: outClusterList}
	err = znp.SendSync(unp.S_ZDO, 0x20, req, &rsp)
	return
}

// request an End Device Bind with the destination device.
func (znp *Znp) ZdoBindReq(dstAddr string, srcAddress string, srcEndpoint zigbee.EndpointId, clusterId uint16,
	dstAddrMode AddrMode, dstAddress string, dstEndpoint zigbee.EndpointId) (rsp *StatusResponse, err error) {
	req := &ZdoBindUnbindReq{DstAddr: dstAddr, SrcAddress: srcAddress, SrcEndpoint: srcEndpoint, ClusterID: clusterId,
		DstAddrMode: dstAddrMode, DstAddress: dstAddress, DstEndpoint: dstEndpoint}
	err = znp.SendSync(unp.S_ZDO, 0x21, req, &rsp)
	return
}

// request a un-bind.
func (znp *Znp) ZdoUnbindReq(dstAddr string, srcAddress string, srcEndpoint zigbee.EndpointId, clusterId uint16,
	dstAddrMode AddrMode, dstAddress string, dstEndpoint zigbee.EndpointId) (rsp *StatusResponse, err error) {
	req := &ZdoBindUnbindReq{DstAddr: dstAddr, SrcAddress: srcAddress, SrcEndpoint: srcEndpoint, ClusterID: clusterId,
		DstAddrMode: dstAddrMode, DstAddress: dstAddress, DstEndpoint: dstEndpoint}
	err = znp.SendSync(unp.S_ZDO, 0x22, req, &rsp)
	return
}

// request the destination device to perform a network discovery
func (znp *Znp) ZdoMgmtNwkDiskReq(dstAddr string, scanChannels *Channels, scanDuration uint8, startIndex uint8) (rsp *StatusResponse, err error) {
	req := &ZdoMgmtNwkDiskReq{DstAddr: dstAddr, ScanChannels: scanChannels, ScanDuration: scanDuration, StartIndex: startIndex}
	err = znp.SendSync(unp.S_ZDO, 0x30, req, &rsp)
	return
}

// request the destination device to perform a LQI query of other
// devices in the network.
func (znp *Znp) ZdoMgmtLqiReq(dstAddr string, startIndex uint8) (rsp *StatusResponse, err error) {
	req := &ZdoMgmtLqiReq{DstAddr: dstAddr, StartIndex: startIndex}
	err = znp.SendSync(unp.S_ZDO, 0x31, req, &rsp)
	return
}

// request the Routing Table of the destination device
func (znp *Znp) ZdoMgmtRtgReq(dstAddr string, startIndex uint8) (rsp *StatusResponse, err error) {
	req := &ZdoMgmtRtgReq{DstAddr: dstAddr, StartIndex: startIndex}
	err = znp.SendSync(unp.S_ZDO, 0x32, req, &rsp)
	return
}

// request the Binding Table of the destination device.
func (znp *Znp) ZdoMgmtBindReq(dstAddr string, startIndex uint8) (rsp *StatusResponse, err error) {
	req := &ZdoMgmtBindReq{DstAddr: dstAddr, StartIndex: startIndex}
	err = znp.SendSync(unp.S_ZDO, 0x33, req, &rsp)
	return
}

// request a Management Leave Request for the target device
func (znp *Znp) ZdoMgmtLeaveReq(dstAddr string, deviceAddr string, removeChildrenRejoin *RemoveChildrenRejoin) (rsp *StatusResponse, err error) {
	req := &ZdoMgmtLeaveReq{DstAddr: dstAddr, DeviceAddr: deviceAddr, RemoveChildrenRejoin: removeChildrenRejoin}
	err = znp.SendSync(unp.S_ZDO, 0x34, req, &rsp)
	return
}

// request the Management Direct Join Request of a designated
// device.
func (znp *Znp) ZdoMgmtDirectJoinReq(dstAddr string, deviceAddr string, capInfo *CapInfo) (rsp *StatusResponse, err error) {
	req := &ZdoMgmtDirectJoinReq{DstAddr: dstAddr, DeviceAddr: deviceAddr, CapInfo: capInfo}
	err = znp.SendSync(unp.S_ZDO, 0x35, req, &rsp)
	return
}

// set the Permit Join for the destination device.
func (znp *Znp) ZdoMgmtPermitJoinReq(addrMode AddrMode, dstAddr string, duration uint8, tcSignificance uint8) (rsp *StatusResponse, err error) {
	req := &ZdoMgmtPermitJoinReq{AddrMode: addrMode, DstAddr: dstAddr, Duration: duration, TCSignificance: tcSignificance}
	err = znp.SendSync(unp.S_ZDO, 0x36, req, &rsp)
	return
}

// is provided to allow updating of network configuration parameters or to request
// information from devices on network conditions in the local operating environment.
func (znp *Znp) ZdoMgmtNwkUpdateReq(dstAddr string, dstAddrMode AddrMode, channelMask *Channels, scanDuration uint8) (rsp *StatusResponse, err error) {
	req := &ZdoMgmtNwkUpdateReq{DstAddr: dstAddr, DstAddrMode: dstAddrMode, ChannelMask: channelMask, ScanDuration: scanDuration}
	err = znp.SendSync(unp.S_ZDO, 0x37, req, &rsp)
	return
}

// registers for a ZDO callback (see reference [3], “6. ZDO Message Requests” for
// example usage).
func (znp *Znp) ZdoMsgCbRegister(clusterId uint16) (rsp *StatusResponse, err error) {
	req := &ZdoMsgCbRegister{ClusterID: clusterId}
	err = znp.SendSync(unp.S_ZDO, 0x3E, req, &rsp)
	return
}

// removes a registration for a ZDO callback (see reference [3], “6. ZDO Message
// Requests” for example usage).
func (znp *Znp) ZdoMsgCbRemove(clusterId uint16) (rsp *StatusResponse, err error) {
	req := &ZdoMsgCbRemove{ClusterID: clusterId}
	err = znp.SendSync(unp.S_ZDO, 0x3F, req, &rsp)
	return
}

// starts the device in the network.
func (znp *Znp) ZdoStartupFromApp(startDelay uint16) (rsp *ZdoStartupFromAppResponse, err error) {
	req := &ZdoStartupFromApp{StartDelay: startDelay}
	err = znp.SendSync(unp.S_ZDO, 0x40, req, &rsp)
	return
}

// starts the device in the network.
func (znp *Znp) ZdoSetLinkKey(shortAddr string, ieeeAddr string, linkKeyData [16]uint8) (rsp *StatusResponse, err error) {
	req := &ZdoSetLinkKey{ShortAddr: shortAddr, IEEEAddr: ieeeAddr, LinkKeyData: linkKeyData}
	err = znp.SendSync(unp.S_ZDO, 0x23, req, &rsp)
	return
}

// removes the application link key of a given device.
func (znp *Znp) ZdoRemoveLinkKey(ieeeAddr string) (rsp *StatusResponse, err error) {
	req := &ZdoRemoveLinkKey{IEEEAddr: ieeeAddr}
	err = znp.SendSync(unp.S_ZDO, 0x24, req, &rsp)
	return
}

// retrieves the application link key of a given device.
func (znp *Znp) ZdoGetLinkKey(ieeeAddr string) (rsp *ZdoGetLinkKeyResponse, err error) {
	req := &ZdoGetLinkKey{IEEEAddr: ieeeAddr}
	err = znp.SendSync(unp.S_ZDO, 0x25, req, &rsp)
	return
}

// initiate a network discovery (active scan).
// Strange response SecOldFrmCount(0xa1)
func (znp *Znp) ZdoNwkDiscoveryReq(scanChannels *Channels, scanDuration uint8) (rsp *StatusResponse, err error) {
	req := &ZdoNwkDiscoveryReq{ScanChannels: scanChannels, ScanDuration: scanDuration}
	err = znp.SendSync(unp.S_ZDO, 0x26, req, &rsp)
	return
}

// request the device to join itself to a parent device on a network.
func (znp *Znp) ZdoJoinReq(logicalChannel uint8, panId uint16, extendedPanId uint64,
	chosenParent string, parentDepth uint8, stackProfile uint8) (rsp *StatusResponse, err error) {
	req := &ZdoJoinReq{LogicalChannel: logicalChannel, PanID: panId, ExtendedPanID: extendedPanId,
		ChosenParent: chosenParent, ParentDepth: parentDepth, StackProfile: stackProfile}
	err = znp.SendSync(unp.S_ZDO, 0x27, req, &rsp)
	return
}

// set rejoin backoff duration and rejoin scan duration for an end device
func (znp *Znp) ZdoSetRejoinParameters(backoffDuration uint32, scanDuration uint32) (rsp *StatusResponse, err error) {
	req := &ZdoSetRejoinParameters{BackoffDuration: backoffDuration, ScanDuration: scanDuration}
	err = znp.SendSync(unp.S_ZDO, 0xCC, req, &rsp)
	return
}

// handles the ZDO security add link key extension message.
func (znp *Znp) ZdoSecAddLinkKey(shortAddress string, extendedAddress string, key [16]uint8) (rsp *StatusResponse, err error) {
	req := &ZdoSecAddLinkKey{ShortAddress: shortAddress, ExtendedAddress: extendedAddress, Key: key}
	err = znp.SendSync(unp.S_ZDO, 0x42, req, &rsp)
	return
}

// handles the ZDO security entry lookup extended extension message
func (znp *Znp) ZdoSecEntryLookupExt(extendedAddress string, entry [5]uint8) (rsp *ZdoSecEntryLookupExtResponse, err error) {
	req := &ZdoSecEntryLookupExt{ExtendedAddress: extendedAddress, Entry: entry}
	err = znp.SendSync(unp.S_ZDO, 0x43, req, &rsp)
	return
}

// handles the ZDO security remove device extended extension message.
func (znp *Znp) ZdoSecDeviceRemove(extendedAddress string) (rsp *StatusResponse, err error) {
	req := &ZdoSecDeviceRemove{ExtendedAddress: extendedAddress}
	err = znp.SendSync(unp.S_ZDO, 0x44, req, &rsp)
	return
}

// handles the ZDO route discovery extension message.
func (znp *Znp) ZdoExtRouteDisc(destinationAddress string, options uint8, radius uint8) (rsp *StatusResponse, err error) {
	req := &ZdoExtRouteDisc{DestinationAddress: destinationAddress, Options: options, Radius: radius}
	err = znp.SendSync(unp.S_ZDO, 0x45, req, &rsp)
	return
}

// handles the ZDO route check extension message.
func (znp *Znp) ZdoExtRouteCheck(destinationAddress string, rtStatus uint8, options uint8) (rsp *StatusResponse, err error) {
	req := &ZdoExtRouteCheck{DestinationAddress: destinationAddress, RTStatus: rtStatus, Options: options}
	err = znp.SendSync(unp.S_ZDO, 0x46, req, &rsp)
	return
}

// handles the ZDO extended remove group extension message.
func (znp *Znp) ZdoExtRemoveGroup(endpoint zigbee.EndpointId, groupId uint16) (rsp *StatusResponse, err error) {
	req := &ZdoExtRemoveGroup{Endpoint: endpoint, GroupID: groupId}
	err = znp.SendSync(unp.S_ZDO, 0x47, req, &rsp)
	return
}

// handles the ZDO extended remove all group extension message.
func (znp *Znp) ZdoExtRemoveAllGroup(endpoint zigbee.EndpointId) (rsp *StatusResponse, err error) {
	req := &ZdoExtRemoveAllGroup{Endpoint: endpoint}
	err = znp.SendSync(unp.S_ZDO, 0x48, req, &rsp)
	return
}

// handles the ZDO extension find all groups for endpoint message
func (znp *Znp) ZdoExtFindAllGroupsEndpoint(endpoint zigbee.EndpointId, groupList []uint16) (rsp *ZdoExtFindAllGroupsEndpointResponse, err error) {
	req := &ZdoExtFindAllGroupsEndpoint{Endpoint: endpoint, GroupList: groupList}
	err = znp.SendSync(unp.S_ZDO, 0x49, req, &rsp)
	return
}

// handles the ZDO extension find all groups for endpoint message
func (znp *Znp) ZdoExtFindGroup(endpoint zigbee.EndpointId, groupID uint16) (rsp *ZdoExtFindGroupResponse, err error) {
	req := &ZdoExtFindGroup{Endpoint: endpoint, GroupID: groupID}
	err = znp.SendSync(unp.S_ZDO, 0x4A, req, &rsp)
	return
}

// handles the ZDO extension add group message.
func (znp *Znp) ZdoExtAddGroup(endpoint zigbee.EndpointId, groupID uint16, groupName string) (rsp *StatusResponse, err error) {
	req := &ZdoExtAddGroup{Endpoint: endpoint, GroupID: groupID, GroupName: groupName}
	err = znp.SendSync(unp.S_ZDO, 0x4B, req, &rsp)
	return
}

// handles the ZDO extension count all groups message.
func (znp *Znp) ZdoExtCountAllGroups() (rsp *ZdoExtCountAllGroupsResponse, err error) {
	err = znp.SendSync(unp.S_ZDO, 0x4C, nil, &rsp)
	return
}

// handles the ZDO extension Get/Set RxOnIdle to ZMac message
func (znp *Znp) ZdoExtRxIdle(setFlag uint8, setValue uint8) (rsp *StatusResponse, err error) { //very unclear from the docs and the code
	req := &ZdoExtRxIdle{SetFlag: setFlag, SetValue: setValue}
	err = znp.SendSync(unp.S_ZDO, 0x4D, req, &rsp)
	return
}

// handles the ZDO security update network key extension message.
func (znp *Znp) ZdoExtUpdateNwkKey(destinationAddress string, keySeqNum uint8, key [128]uint8) (rsp *StatusResponse, err error) {
	req := &ZdoExtUpdateNwkKey{DestinationAddress: destinationAddress, KeySeqNum: keySeqNum, Key: key}
	err = znp.SendSync(unp.S_ZDO, 0x4E, req, &rsp)
	return
}

// handles the ZDO security switch network key extension message.
func (znp *Znp) ZdoExtSwitchNwkKey(destinationAddress string, keySeqNum uint8) (rsp *StatusResponse, err error) {
	req := &ZdoExtSwitchNwkKey{DestinationAddress: destinationAddress, KeySeqNum: keySeqNum}
	err = znp.SendSync(unp.S_ZDO, 0x4F, req, &rsp)
	return
}

// handles the ZDO extension network message.
func (znp *Znp) ZdoExtNwkInfo() (rsp *ZdoExtNwkInfoResponse, err error) {
	err = znp.SendSync(unp.S_ZDO, 0x50, nil, &rsp)
	return
}

// handles the ZDO extension Security Manager APS Remove Request message.
func (znp *Znp) ZdoExtSeqApsRemoveReq(nwkAddress string, extendedAddress string, parentAddress string) (rsp *StatusResponse, err error) {
	req := &ZdoExtSeqApsRemoveReq{NwkAddress: nwkAddress, ExtendedAddress: extendedAddress, ParentAddress: parentAddress}
	err = znp.SendSync(unp.S_ZDO, 0x51, req, &rsp)
	return
}

// forces a network concentrator change by resetting zgConcentratorEnable and
// zgConcentratorDiscoveryTime from NV and set nwk event.
func (znp *Znp) ZdoForceConcentratorChange() {
	znp.SendAsync(unp.S_ZDO, 0x52, nil, nil)
}

// set parameters not settable through NV.
func (znp *Znp) ZdoExtSetParams(useMulticast uint8) (rsp *StatusResponse, err error) {
	req := &ZdoExtSetParams{UseMulticast: useMulticast}
	err = znp.SendSync(unp.S_ZDO, 0x53, req, &rsp)
	return
}

// handles ZDO network address of interest request.
func (znp *Znp) ZdoNwkAddrOfInterestReq(destAddr string, nwkAddrOfInterest string, cmd uint8) (rsp *StatusResponse, err error) {
	req := &ZdoNwkAddrOfInterestReq{DestAddr: destAddr, NwkAddrOfInterest: nwkAddrOfInterest, Cmd: cmd}
	err = znp.SendSync(unp.S_ZDO, 0x29, req, &rsp)
	return
}

// =======APP_CNF=======

// sets the network frame counter to the value specified in the Frame Counter Value.
// For projects with multiple instances of frame counter, the message sets the frame counter of the
// current network.
func (znp *Znp) AppCnfSetNwkFrameCounter(frameCounterValue uint8) (rsp *StatusResponse, err error) {
	req := &AppCnfSetNwkFrameCounter{FrameCounterValue: frameCounterValue}
	err = znp.SendSync(unp.S_APP_CNF, 0xFF, req, &rsp)
	return
}

// sets the default value used by parent device to expire legacy child devices.
func (znp *Znp) AppCnfSetDefaultEndDeviceTimeout(timeout Timeout) (rsp *StatusResponse, err error) {
	req := &AppCnfSetDefaultEndDeviceTimeout{Timeout: timeout}
	err = znp.SendSync(unp.S_APP_CNF, 0x01, req, &rsp)
	return
}

// sets in ZED the timeout value to be send to parent device for child expiring.
func (znp *Znp) AppCnfSetEndDeviceTimeout(timeout Timeout) (rsp *StatusResponse, err error) {
	req := &AppCnfSetEndDeviceTimeout{Timeout: timeout}
	err = znp.SendSync(unp.S_APP_CNF, 0x02, req, &rsp)
	return
}

// sets the AllowRejoin TC policy.
func (znp *Znp) AppCnfSetAllowRejoinTcPolicy(allowRejoin uint8) (rsp *StatusResponse, err error) {
	req := &AppCnfSetAllowRejoinTcPolicy{AllowRejoin: allowRejoin}
	err = znp.SendSync(unp.S_APP_CNF, 0x03, req, &rsp)
	return
}

// set the commissioning methods to be executed. Initialization of BDB is executed with this call,
// regardless of its parameters.
func (znp *Znp) AppCnfBdbStartCommissioning(commissioningMode CommissioningMode) (rsp *StatusResponse, err error) {
	req := &AppCnfBdbStartCommissioning{CommissioningMode: commissioningMode}
	err = znp.SendSync(unp.S_APP_CNF, 0x05, req, &rsp)
	return
}

// sets  BDB primary or secondary channel masks.
func (znp *Znp) AppCnfBdbSetChannel(isPrimary uint8, channel *Channels) (rsp *StatusResponse, err error) {
	req := &AppCnfBdbSetChannel{IsPrimary: isPrimary, Channel: channel}
	err = znp.SendSync(unp.S_APP_CNF, 0x08, req, &rsp)
	return
}

// add a preconfigured key (plain key or IC) to Trust Center device.
func (znp *Znp) AppCnfBdbAddInstallCode(installCodeFormat InstallCodeFormat, ieeeAddr string, installCode []uint8) (rsp *StatusResponse, err error) {
	req := &AppCnfBdbAddInstallCode{InstallCodeFormat: installCodeFormat, IEEEAddr: ieeeAddr, InstallCode: installCode}
	err = znp.SendSync(unp.S_APP_CNF, 0x04, req, &rsp)
	return
}

// sets the policy flag on Trust Center device to mandate or not the TCLK exchange procedure.
func (znp *Znp) AppCnfBdbSetTcRequireKeyExchange(bdbTrustCenterRequireKeyExchange uint8) (rsp *StatusResponse, err error) {
	req := &AppCnfBdbSetTcRequireKeyExchange{BdbTrustCenterRequireKeyExchange: bdbTrustCenterRequireKeyExchange}
	err = znp.SendSync(unp.S_APP_CNF, 0x09, req, &rsp)
	return
}

// sets the policy to mandate or not the usage of an Install Code upon joining.
func (znp *Znp) AppCnfBdbSetJoinUsesInstallCodeKey(bdbJoinUsesInstallCodeKey uint8) (rsp *StatusResponse, err error) {
	req := &AppCnfBdbSetJoinUsesInstallCodeKey{BdbJoinUsesInstallCodeKey: bdbJoinUsesInstallCodeKey}
	err = znp.SendSync(unp.S_APP_CNF, 0x06, req, &rsp)
	return
}

// on joining devices, set the default key or an install code to attempt to join the network.
func (znp *Znp) AppCnfBdbSetActiveDefaultCentralizedKey(useGlobal uint8, installCode [18]uint8) (rsp *StatusResponse, err error) {
	req := &AppCnfBdbSetActiveDefaultCentralizedKey{UseGlobal: useGlobal, InstallCode: installCode}
	err = znp.SendSync(unp.S_APP_CNF, 0x07, req, &rsp)
	return
}

// instruct the ZED to try to rejoin its previews network. Use only in ZED devices.
func (znp *Znp) AppCnfBdbZedAttemptRecoverNwk(useGlobal uint8, installCode [18]uint8) (rsp *StatusResponse, err error) {
	err = znp.SendSync(unp.S_APP_CNF, 0x0A, nil, &rsp)
	return
}

// =======GP=======

// callback to receive notifications from BDB process.
func (znp *Znp) GpDataReq(action GpAction, txOptions *TxOptions, applicationId uint8, srcId uint32,
	gpdIEEEAddress string, endpoint zigbee.EndpointId, gpdCommandId uint8, gpdasdu []uint8,
	gpepHandle uint8, gpTxQueueEntryLifetime uint32) (rsp *StatusResponse, err error) {
	req := &GpDataReq{Action: action, TxOptions: txOptions, ApplicationID: applicationId,
		SrcID: srcId, GPDIEEEAddress: gpdIEEEAddress, Endpoint: endpoint,
		GPDCommandID: gpdCommandId, GPDASDU: gpdasdu, GPEPHandle: gpepHandle,
		GPTxQueueEntryLifetime: gpTxQueueEntryLifetime}
	err = znp.SendSync(unp.S_GP, 0x01, req, &rsp)
	return
}

// provides a mechanism for the Green Power EndPoint to provide security data into
// the dGP stub.
func (znp *Znp) GpSecRsp(status GpStatus, dGPStubHandle uint8, applicationID uint8, srcID uint32,
	gpdIEEEAddress string, endpoint zigbee.EndpointId, gpdFSecurityLevel uint8, gpdFKeyType uint8,
	gpdKey [16]uint8, gpdSecurityFrameCounter uint32) (rsp *StatusResponse, err error) {
	req := &GpSecRsp{Status: status, DGPStubHandle: dGPStubHandle, ApplicationID: applicationID,
		SrcID: srcID, GPDIEEEAddress: gpdIEEEAddress, Endpoint: endpoint, GPDFSecurityLevel: gpdFSecurityLevel,
		GPDFKeyType: gpdFKeyType, GPDKey: gpdKey, GPDSecurityFrameCounter: gpdSecurityFrameCounter}
	err = znp.SendSync(unp.S_GP, 0x02, req, &rsp)
	return
}

// dynamically new's you an async command struct from *asyncCommandAllocators*
func NewConcreteAsyncCommand(subsystem unp.Subsystem, command byte) (interface{}, error) {
	key := acKey{subsystem, command}
	alloc, found := asyncCommandAllocators[key]
	if !found {
		return nil, fmt.Errorf("unknown async command received: %d/%d", subsystem, command)
	}

	// in practice returns &AfDataConfirm{} but chooses datatype dynamically based on subsys + cmd
	return alloc(), nil
}

// async command key (shortened to reduce repetition below)
type acKey struct {
	subsystem unp.Subsystem
	command   byte
}

const (
	AfIncomingMessageId = 0x81 // exported b/c used in tests
)

var asyncCommandAllocators = map[acKey]func() interface{}{
	//AF
	acKey{unp.S_AF, 0x80}:                func() interface{} { return &AfDataConfirm{} },
	acKey{unp.S_AF, 0x83}:                func() interface{} { return &AfReflectError{} },
	acKey{unp.S_AF, AfIncomingMessageId}: func() interface{} { return &AfIncomingMessage{} },
	acKey{unp.S_AF, 0x82}:                func() interface{} { return &AfIncomingMessageExt{} },

	//DEBUG
	acKey{unp.S_DBG, 0x00}: func() interface{} { return &DebugMsg{} },

	//SAPI
	acKey{unp.S_SAPI, 0x80}: func() interface{} { return &SapiZbStartConfirm{} },
	acKey{unp.S_SAPI, 0x81}: func() interface{} { return &SapiZbBindConfirm{} },
	acKey{unp.S_SAPI, 0x82}: func() interface{} { return &SapiZbAllowBindConfirm{} },
	acKey{unp.S_SAPI, 0x83}: func() interface{} { return &SapiZbSendDataConfirm{} },
	acKey{unp.S_SAPI, 0x87}: func() interface{} { return &SapiZbReceiveDataIndication{} },
	acKey{unp.S_SAPI, 0x85}: func() interface{} { return &SapiZbFindDeviceConfirm{} },

	//SYS
	acKey{unp.S_SYS, 0x80}: func() interface{} { return &SysResetInd{} },
	acKey{unp.S_SYS, 0x81}: func() interface{} { return &SysOsalTimerExpired{} },

	//UTIL
	acKey{unp.S_UTIL, 0xE0}: func() interface{} { return &UtilSyncReq{} },
	acKey{unp.S_UTIL, 0xE1}: func() interface{} { return &UtilZclKeyEstablishInd{} },

	//ZDO
	acKey{unp.S_ZDO, 0x80}: func() interface{} { return &ZdoNwkAddrRsp{} },
	acKey{unp.S_ZDO, 0x81}: func() interface{} { return &ZdoIEEEAddrRsp{} },
	acKey{unp.S_ZDO, 0x82}: func() interface{} { return &ZdoNodeDescRsp{} },
	acKey{unp.S_ZDO, 0x83}: func() interface{} { return &ZdoPowerDescRsp{} },
	acKey{unp.S_ZDO, 0x84}: func() interface{} { return &ZdoSimpleDescRsp{} },
	acKey{unp.S_ZDO, 0x85}: func() interface{} { return &ZdoActiveEpRsp{} },
	acKey{unp.S_ZDO, 0x86}: func() interface{} { return &ZdoMatchDescRsp{} },
	acKey{unp.S_ZDO, 0x87}: func() interface{} { return &ZdoComplexDescRsp{} },
	acKey{unp.S_ZDO, 0x88}: func() interface{} { return &ZdoUserDescRsp{} },
	acKey{unp.S_ZDO, 0x89}: func() interface{} { return &ZdoUserDescConf{} },
	acKey{unp.S_ZDO, 0x8A}: func() interface{} { return &ZdoServerDiscRsp{} },
	acKey{unp.S_ZDO, 0xA0}: func() interface{} { return &ZdoEndDeviceBindRsp{} },
	acKey{unp.S_ZDO, 0xA1}: func() interface{} { return &ZdoBindRsp{} },
	acKey{unp.S_ZDO, 0xA2}: func() interface{} { return &ZdoUnbindRsp{} },
	acKey{unp.S_ZDO, 0xB0}: func() interface{} { return &ZdoMgmtNwkDiscRsp{} },
	acKey{unp.S_ZDO, 0xB1}: func() interface{} { return &ZdoMgmtLqiRsp{} },
	acKey{unp.S_ZDO, 0xB2}: func() interface{} { return &ZdoMgmtRtgRsp{} },
	acKey{unp.S_ZDO, 0xB3}: func() interface{} { return &ZdoMgmtBindRsp{} },
	acKey{unp.S_ZDO, 0xB4}: func() interface{} { return &ZdoMgmtLeaveRsp{} },
	acKey{unp.S_ZDO, 0xB5}: func() interface{} { return &ZdoMgmtDirectJoinRsp{} },
	acKey{unp.S_ZDO, 0xB6}: func() interface{} { return &ZdoMgmtPermitJoinRsp{} },
	acKey{unp.S_ZDO, 0xC0}: func() interface{} { return &ZdoStateChangeInd{} },
	acKey{unp.S_ZDO, 0xC1}: func() interface{} { return &ZdoEndDeviceAnnceInd{} },
	acKey{unp.S_ZDO, 0xC2}: func() interface{} { return &ZdoMatchDescRpsSent{} },
	acKey{unp.S_ZDO, 0xC3}: func() interface{} { return &ZdoStatusErrorRsp{} },
	acKey{unp.S_ZDO, 0xC4}: func() interface{} { return &ZdoSrcRtgInd{} },
	acKey{unp.S_ZDO, 0xC5}: func() interface{} { return &ZdoBeaconNotifyInd{} },
	acKey{unp.S_ZDO, 0xC6}: func() interface{} { return &ZdoJoinCnf{} },
	acKey{unp.S_ZDO, 0xC7}: func() interface{} { return &ZdoNwkDiscoveryCnf{} },
	acKey{unp.S_ZDO, 0xC9}: func() interface{} { return &ZdoLeaveInd{} },
	acKey{unp.S_ZDO, 0xFF}: func() interface{} { return &ZdoMsgCbIncoming{} },
	acKey{unp.S_ZDO, 0xCA}: func() interface{} { return &ZdoTcDevInd{} },
	acKey{unp.S_ZDO, 0xCB}: func() interface{} { return &ZdoPermitJoinInd{} },

	//APP
	acKey{unp.S_APP_CNF, 0x80}: func() interface{} { return &AppCnfBdbCommissioningNotification{} },

	//GP
	acKey{unp.S_GP, 0x01}: func() interface{} { return &GpDataReq{} },
	acKey{unp.S_GP, 0x02}: func() interface{} { return &GpSecRsp{} },
	acKey{unp.S_GP, 0x05}: func() interface{} { return &GpDataCnf{} },
	acKey{unp.S_GP, 0x03}: func() interface{} { return &GpSecReq{} },
	acKey{unp.S_GP, 0x04}: func() interface{} { return &GpDataInd{} },
}
