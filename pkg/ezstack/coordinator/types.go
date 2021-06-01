package coordinator

import (
	"bytes"
	"fmt"
	"reflect"

	. "github.com/function61/hautomo/pkg/builtin"
	"github.com/function61/hautomo/pkg/ezstack/zigbee"
	"github.com/function61/hautomo/pkg/ezstack/znp"
)

type Configuration struct {
	NetworkConfiguration
	Led    bool `json:"LED,omitempty"`
	Serial *Serial
}

type NetworkConfiguration struct {
	IEEEAddress string // has prefix "0x"
	PanId       zigbee.PANID
	ExtPanId    zigbee.ExtendedPANID
	NetworkKey  []byte
	Channel     uint8
}

func (c NetworkConfiguration) Equal(other NetworkConfiguration) bool {
	if c.IEEEAddress != other.IEEEAddress {
		return false
	}

	if c.PanId != other.PanId {
		return false
	}

	if c.ExtPanId != other.ExtPanId {
		return false
	}

	if !bytes.Equal(c.NetworkKey, other.NetworkKey) {
		return false
	}

	if c.Channel != other.Channel {
		return false
	}

	return true
}

func (c *Configuration) GetNetworkKey() (zigbee.NetworkKey, error) {
	key := zigbee.NetworkKey{}
	if copy(key[:], c.NetworkKey) != len(key) {
		return key, fmt.Errorf("invalid length for NetworkKey: %d", len(c.NetworkKey))
	}

	return key, nil
}

func (c *Configuration) Valid() error {
	errorIfEmpty := func(empty bool, field string) error {
		if empty {
			return fmt.Errorf("required: %s", field)
		} else {
			return nil
		}
	}

	return FirstError(
		errorIfEmpty(c.IEEEAddress == "", "IEEEAddress"),
		errorIfEmpty(c.PanId == 0, "PanId"),
		errorIfEmpty(c.ExtPanId == 0, "ExtPanId"),
		errorIfEmpty(c.Channel == 0, "Channel"),
		errorIfEmpty(len(c.NetworkKey) != 16, "NetworkKey"),
		errorIfEmpty(c.Serial == nil, "Serial"),
	)

	return nil
}

type Serial struct {
	Port     string
	BaudRate *int `json:"BaudRate,omitempty"` // if nil, optimal default is used
}

func (s Serial) BaudRateOrDefault() int {
	if s.BaudRate != nil {
		return *s.BaudRate
	} else {
		// confirmed by zigbee2mqtt docs, TI forums & OpenHab forums
		return 115200
	}
}

var SysResetIndType = reflect.TypeOf(&znp.SysResetInd{})
var ZdoActiveEpRspType = reflect.TypeOf(&znp.ZdoActiveEpRsp{})
var ZdoSimpleDescRspType = reflect.TypeOf(&znp.ZdoSimpleDescRsp{})
var ZdoNodeDescRspType = reflect.TypeOf(&znp.ZdoNodeDescRsp{})
var ZdoBindRspType = reflect.TypeOf(&znp.ZdoBindRsp{})
var ZdoUnbindRspType = reflect.TypeOf(&znp.ZdoUnbindRsp{})
