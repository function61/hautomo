package znp

import (
	"fmt"

	"github.com/function61/hautomo/pkg/ezstack/binstruct"
	"github.com/function61/hautomo/pkg/ezstack/zigbee"
)

// Manages NVRAM (= stored parameters) of CC2531 stick

type NVRAMItemId uint16

type NVRAMItem interface {
	// returns the item ID ("slot ID") where this particular parameter is stored in the NVRAM.
	// used as ID for reads / writes.
	ItemID() NVRAMItemId
}

func (z *Znp) NVRAMRead(item NVRAMItem) error {
	resp, err := (&SysOsalNvRead{
		ID: item.ItemID(),
	}).Send(z)
	if err != nil {
		return fmt.Errorf("NVRAMRead: %d: %w", item.ItemID(), err)
	}

	return binstruct.Decode(resp.Value, item)
}

func (z *Znp) NVRAMWrite(item NVRAMItem) error {
	resp, err := (&SysOsalNvWrite{
		ID:    item.ItemID(),
		Value: binstruct.Encode(item),
	}).Send(z)
	if err != nil {
		return fmt.Errorf("NVRAMWrite: %d: %w", item.ItemID(), err)
	}

	return resp.Status.Error()
}

type ZCDNVStartUpOption struct {
	StartOption uint8 // is supposed to be 0x03, though I don't know what it means
}

func (n *ZCDNVStartUpOption) ItemID() NVRAMItemId {
	return 0x0003
}

type ZCDNVLogicalType struct {
	LogicalType zigbee.LogicalType
}

func (n *ZCDNVLogicalType) ItemID() NVRAMItemId {
	return 0x0087
}

// probably related: https://www.digi.com/resources/documentation/Digidocs/90001942-13/concepts/c_zb_security_model.htm?TocPath=Security+and+encryption%7CZigBee+security+model%7C_____0
type ZCDNVSecurityMode struct {
	Enabled uint8
}

func (n *ZCDNVSecurityMode) ItemID() NVRAMItemId {
	return 0x0064
}

type ZCDNVPreCfgKeysEnable struct {
	Enabled uint8 // whether to enable use of key specified in *ZCDNVPreCfgKey*?
}

func (n *ZCDNVPreCfgKeysEnable) ItemID() NVRAMItemId {
	return 0x0063
}

type ZCDNVPreCfgKey struct {
	NetworkKey zigbee.NetworkKey
}

func (n *ZCDNVPreCfgKey) ItemID() NVRAMItemId {
	return 0x0062
}

// what does this mean? Zigbee Device Object something something?
type ZCDNVZDODirectCB struct {
	Enabled uint8 // supposed to be enabled, but what does it do?
}

func (n *ZCDNVZDODirectCB) ItemID() NVRAMItemId {
	return 0x008f
}

type ZCDNVChanList struct {
	Channels [4]byte
}

func (n *ZCDNVChanList) ItemID() NVRAMItemId {
	return 0x0084
}

type ZCDNVPANID struct {
	PANID zigbee.PANID
}

func (n *ZCDNVPANID) ItemID() NVRAMItemId {
	return 0x0083
}

type ZCDNVExtPANID struct {
	ExtendedPANID zigbee.ExtendedPANID
}

func (n *ZCDNVExtPANID) ItemID() NVRAMItemId {
	return 0x002d
}

type ZCDNVUseDefaultTCLK struct {
	Enabled uint8 // supposed to be required for ZStack < 3.x
}

func (n *ZCDNVUseDefaultTCLK) ItemID() NVRAMItemId {
	return 0x006d
}
