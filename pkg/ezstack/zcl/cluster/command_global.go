package cluster

// generic commands can be thought of device-level commands, like (un)binding

import (
	"encoding/binary"
	"io"
	"strconv"

	"github.com/dyrkin/composer"
	"github.com/function61/hautomo/pkg/ezstack/binstruct"
)

// WARNING: many of these structs have very important order as they directly influence data
//          (un)marshaling

type ReadAttributesCommand struct {
	AttributeIDs []uint16
}

type TimeOfDay struct {
	Hours      uint8
	Minutes    uint8
	Seconds    uint8
	Hundredths uint8
}

type Date struct {
	Year       uint8
	Month      uint8
	DayOfMonth uint8
	DayOfWeek  uint8
}

type Attribute struct {
	DataType ZclDataType
	Value    interface{}
}

type ReadAttributeStatus struct {
	AttributeName string `transient:"true"`
	AttributeID   uint16
	Status        ZclStatus
	Attribute     *Attribute `cond:"uint:Status==0"`
}

type ReadAttributesResponse struct {
	ReadAttributeStatuses []*ReadAttributeStatus
}

type WriteAttributeRecord struct {
	AttributeName string `transient:"true"`
	AttributeID   uint16
	Attribute     *Attribute
}

type WriteAttributesCommand struct {
	WriteAttributeRecords []*WriteAttributeRecord
}

type WriteAttributesUndividedCommand struct {
	WriteAttributeRecords []*WriteAttributeRecord
}

type WriteAttributesNoResponseCommand struct {
	WriteAttributeRecords []*WriteAttributeRecord
}

type WriteAttributeStatus struct {
	Status        ZclStatus
	AttributeName string `transient:"true"`
	AttributeID   uint16
}

type WriteAttributesResponse struct {
	WriteAttributeStatuses []*WriteAttributeStatus
}

type AttributeReportingConfigurationRecord struct {
	Direction                ReportDirection
	AttributeName            string `transient:"true"`
	AttributeID              uint16
	AttributeDataType        ZclDataType `cond:"uint:Direction==0"`
	MinimumReportingInterval uint16      `cond:"uint:Direction==0"`
	MaximumReportingInterval uint16      `cond:"uint:Direction==0"`
	ReportableChange         *Attribute  `cond:"uint:Direction==0"`
	TimeoutPeriod            uint16      `cond:"uint:Direction==1"`
}

type ConfigureReportingCommand struct {
	AttributeReportingConfigurationRecords []*AttributeReportingConfigurationRecord
}

type AttributeStatusRecord struct {
	Status        ZclStatus
	Direction     ReportDirection
	AttributeName string `transient:"true"`
	AttributeID   uint16
}

type ConfigureReportingResponse struct {
	AttributeStatusRecords []*AttributeStatusRecord
}

type AttributeRecord struct {
	Direction     ReportDirection
	AttributeName string `transient:"true"`
	AttributeID   uint16
}

type ReadReportingConfigurationCommand struct {
	AttributeRecords []*AttributeRecord
}

type AttributeReportingConfigurationResponseRecord struct {
	Status                   ZclStatus
	Direction                ReportDirection
	AttributeName            string `transient:"true"`
	AttributeID              uint16
	AttributeDataType        ZclDataType `cond:"uint:Direction==0;uint:Status==0"`
	MinimumReportingInterval uint16      `cond:"uint:Direction==0;uint:Status==0"`
	MaximumReportingInterval uint16      `cond:"uint:Direction==0;uint:Status==0"`
	ReportableChange         *Attribute  `cond:"uint:Direction==0;uint:Status==0"`
	TimeoutPeriod            uint16      `cond:"uint:Direction==1;uint:Status==0"`
}

type ReadReportingConfigurationResponse struct {
	AttributeReportingConfigurationResponseRecords []*AttributeReportingConfigurationResponseRecord
}

type AttributeReport struct {
	AttributeName string `transient:"true"`
	AttributeID   uint16
	Attribute     *Attribute
}

type ReportAttributesCommand struct {
	AttributeReports []*AttributeReport
}

type DefaultResponseCommand struct {
	CommandID uint8
	Status    ZclStatus
}

type DiscoverAttributesCommand struct {
	StartAttributeID            uint16
	MaximumAttributeIdentifiers uint8
}

type AttributeInformation struct {
	AttributeName     string `transient:"true"`
	AttributeID       uint16
	AttributeDataType ZclDataType
}

type DiscoverAttributesResponse struct {
	DiscoveryComplete     uint8
	AttributeInformations []*AttributeInformation
}

type AttributeSelector struct {
	AttributeName string `transient:"true"`
	AttributeID   uint16
	Selector      []uint16 `size:"1"`
}

type ReadAttributesStructuredCommand struct {
	AttributeSelectors []*AttributeSelector
}

type WriteAttributeStructuredRecord struct {
	AttributeName string `transient:"true"`
	AttributeID   uint16
	Selector      []uint16 `size:"1"`
	Attribute     *Attribute
}

type WriteAttributesStructuredCommand struct {
	WriteAttributeStructuredRecords []*WriteAttributeStructuredRecord
}

type WriteAttributeStatusRecord struct {
	Status        ZclStatus
	AttributeName string `transient:"true"`
	AttributeID   uint16
	Selector      []uint16 `size:"1"`
}

type WriteAttributesStructuredResponse struct {
	WriteAttributeStatusRecords []*WriteAttributeStatusRecord
}

type DiscoverCommandsReceivedCommand struct {
	StartCommandID            uint8
	MaximumCommandIdentifiers uint8
}

type DiscoverCommandsReceivedResponse struct {
	DiscoveryComplete  uint8
	CommandIdentifiers []uint8
}

type DiscoverCommandsGeneratedCommand struct {
	StartCommandID            uint8
	MaximumCommandIdentifiers uint8
}

type DiscoverCommandsGeneratedResponse struct {
	DiscoveryComplete  uint8
	CommandIdentifiers []uint8
}

type DiscoverAttributesExtendedCommand struct {
	StartAttributeID            uint16
	MaximumAttributeIdentifiers uint8
}

type AttributeAccessControl struct {
	Readable   uint8 `bits:"0b00000001" bitmask:"start"`
	Writeable  uint8 `bits:"0b00000010"`
	Reportable uint8 `bits:"0b00000100" bitmask:"end"`
}

type ExtendedAttributeInformation struct {
	AttributeName          string `transient:"true"`
	AttributeID            uint16
	AttributeDataType      ZclDataType
	AttributeAccessControl *AttributeAccessControl
}

type DiscoverAttributesExtendedResponse struct {
	DiscoveryComplete             uint8
	ExtendedAttributeInformations []*ExtendedAttributeInformation
}

func (a *Attribute) Serialize(w io.Writer) {
	c := composer.NewWithW(w)
	writeAttribute(c, a.DataType, a.Value)
	c.Flush()
}

func writeAttribute(c *composer.Composer, dataType ZclDataType, value interface{}) {
	c.Uint8(uint8(dataType))
	switch dataType {
	case ZclDataTypeNoData:
	case ZclDataTypeData8:
		b := value.([1]byte)
		c.Bytes(b[:])
	case ZclDataTypeData16:
		b := value.([2]byte)
		c.Bytes(b[:])
	case ZclDataTypeData24:
		b := value.([3]byte)
		c.Bytes(b[:])
	case ZclDataTypeData32:
		b := value.([4]byte)
		c.Bytes(b[:])
	case ZclDataTypeData40:
		b := value.([5]byte)
		c.Bytes(b[:])
	case ZclDataTypeData48:
		b := value.([6]byte)
		c.Bytes(b[:])
	case ZclDataTypeData56:
		b := value.([7]byte)
		c.Bytes(b[:])
	case ZclDataTypeData64:
		b := value.([8]byte)
		c.Bytes(b[:])
	case ZclDataTypeBoolean:
		b := value.(bool)
		if b {
			c.Byte(1)
		} else {
			c.Byte(0)
		}
	case ZclDataTypeBitmap8:
		b := value.(uint64)
		c.Uint(binary.LittleEndian, b, 1)
	case ZclDataTypeBitmap16:
		b := value.(uint64)
		c.Uint(binary.LittleEndian, b, 2)
	case ZclDataTypeBitmap24:
		b := value.(uint64)
		c.Uint(binary.LittleEndian, b, 3)
	case ZclDataTypeBitmap32:
		b := value.(uint64)
		c.Uint(binary.LittleEndian, b, 4)
	case ZclDataTypeBitmap40:
		b := value.(uint64)
		c.Uint(binary.LittleEndian, b, 5)
	case ZclDataTypeBitmap48:
		b := value.(uint64)
		c.Uint(binary.LittleEndian, b, 6)
	case ZclDataTypeBitmap56:
		b := value.(uint64)
		c.Uint(binary.LittleEndian, b, 7)
	case ZclDataTypeBitmap64:
		b := value.(uint64)
		c.Uint(binary.LittleEndian, b, 8)
	case ZclDataTypeUint8:
		b := value.(uint64)
		c.Uint(binary.LittleEndian, b, 1)
	case ZclDataTypeUint16:
		b := value.(uint64)
		c.Uint(binary.LittleEndian, b, 2)
	case ZclDataTypeUint24:
		b := value.(uint64)
		c.Uint(binary.LittleEndian, b, 3)
	case ZclDataTypeUint32:
		b := value.(uint64)
		c.Uint(binary.LittleEndian, b, 4)
	case ZclDataTypeUint40:
		b := value.(uint64)
		c.Uint(binary.LittleEndian, b, 5)
	case ZclDataTypeUint48:
		b := value.(uint64)
		c.Uint(binary.LittleEndian, b, 6)
	case ZclDataTypeUint56:
		b := value.(uint64)
		c.Uint(binary.LittleEndian, b, 7)
	case ZclDataTypeUint64:
		b := value.(uint64)
		c.Uint(binary.LittleEndian, b, 8)
	case ZclDataTypeInt8:
		b := value.(int64)
		c.Int(binary.LittleEndian, b, 1)
	case ZclDataTypeInt16:
		b := value.(int64)
		c.Int(binary.LittleEndian, b, 2)
	case ZclDataTypeInt24:
		b := value.(int64)
		c.Int(binary.LittleEndian, b, 3)
	case ZclDataTypeInt32:
		b := value.(int64)
		c.Int(binary.LittleEndian, b, 4)
	case ZclDataTypeInt40:
		b := value.(int64)
		c.Int(binary.LittleEndian, b, 5)
	case ZclDataTypeInt48:
		b := value.(int64)
		c.Int(binary.LittleEndian, b, 6)
	case ZclDataTypeInt56:
		b := value.(int64)
		c.Int(binary.LittleEndian, b, 7)
	case ZclDataTypeInt64:
		b := value.(int64)
		c.Int(binary.LittleEndian, b, 8)
	case ZclDataTypeEnum8:
		b := value.(uint64)
		c.Uint(binary.LittleEndian, b, 1)
	case ZclDataTypeEnum16:
		b := value.(uint64)
		c.Uint(binary.LittleEndian, b, 2)
	case ZclDataTypeSemiPrec:
	case ZclDataTypeSinglePrec:
	case ZclDataTypeDoublePrec:
	case ZclDataTypeOctetStr:
		b := value.(string)
		c.Uint8(uint8(len(b)))
		c.String(b)
	case ZclDataTypeCharStr:
		b := value.(string)
		c.Uint8(uint8(len(b)))
		c.String(b)
	case ZclDataTypeLongOctetStr:
		b := value.(string)
		c.Uint16le(uint16(len(b)))
		c.String(b)
	case ZclDataTypeLongCharStr:
		b := value.(string)
		c.Uint16le(uint16(len(b)))
		c.String(b)
	case ZclDataTypeArray, ZclDataTypeSet, ZclDataTypeBag:
		attributes := value.([]*Attribute)
		c.Uint16le(uint16(len(attributes)))
		for _, attribute := range attributes {
			writeAttribute(c, attribute.DataType, attribute.Value)
		}
	case ZclDataTypeStruct:
	case ZclDataTypeTod:
		b := value.(*TimeOfDay)
		c.Uint8(b.Hours)
		c.Uint8(b.Minutes)
		c.Uint8(b.Seconds)
		c.Uint8(b.Hundredths)
	case ZclDataTypeDate:
		b := value.(*Date)
		c.Uint8(b.Year)
		c.Uint8(b.Month)
		c.Uint8(b.DayOfMonth)
		c.Uint8(b.DayOfWeek)
	case ZclDataTypeUtc:
		b := value.(uint32)
		c.Uint32le(b)
	case ZclDataTypeClusterId:
		b := value.(uint16)
		c.Uint16le(b)
	case ZclDataTypeAttrId:
		b := value.(uint16)
		c.Uint16le(b)
	case ZclDataTypeBacOid:
		b := value.(uint32)
		c.Uint32le(b)
	case ZclDataTypeIeeeAddr:
		b := value.(string)
		v, _ := strconv.ParseUint(b[2:], 16, 64)
		c.Uint64le(v)
	case ZclDataType_128BitSecKey:
		b := value.([16]byte)
		c.Bytes(b[:])
	case ZclDataTypeUnknown:

	}
}

// WARNING: this is called magically from dyrkin/bin
func (a *Attribute) Deserialize(r io.Reader) {
	c := composer.NewWithR(r)
	a.DataType, a.Value = readAttribute(c)
}

func readAttribute(c *composer.Composer) (dataType ZclDataType, value interface{}) {
	dt, _ := c.ReadByte()
	dataType = ZclDataType(dt)

	switch dataType {
	case ZclDataTypeNoData:
		value = nil
	case ZclDataTypeData8:
		var buf [1]byte
		c.ReadBuf(buf[:])
		value = buf
	case ZclDataTypeData16:
		var buf [2]byte
		c.ReadBuf(buf[:])
		value = buf
	case ZclDataTypeData24:
		var buf [3]byte
		c.ReadBuf(buf[:])
		value = buf
	case ZclDataTypeData32:
		var buf [4]byte
		c.ReadBuf(buf[:])
		value = buf
	case ZclDataTypeData40:
		var buf [5]byte
		c.ReadBuf(buf[:])
		value = buf
	case ZclDataTypeData48:
		var buf [6]byte
		c.ReadBuf(buf[:])
		value = buf
	case ZclDataTypeData56:
		var buf [7]byte
		c.ReadBuf(buf[:])
		value = buf
	case ZclDataTypeData64:
		var buf [8]byte
		c.ReadBuf(buf[:])
		value = buf
	case ZclDataTypeBoolean:
		b, _ := c.ReadByte()
		value = b > 0
	case ZclDataTypeBitmap8:
		value = c.ReadUint(binary.LittleEndian, 1)
	case ZclDataTypeBitmap16:
		value = c.ReadUint(binary.LittleEndian, 2)
	case ZclDataTypeBitmap24:
		value = c.ReadUint(binary.LittleEndian, 3)
	case ZclDataTypeBitmap32:
		value = c.ReadUint(binary.LittleEndian, 4)
	case ZclDataTypeBitmap40:
		value = c.ReadUint(binary.LittleEndian, 5)
	case ZclDataTypeBitmap48:
		value = c.ReadUint(binary.LittleEndian, 6)
	case ZclDataTypeBitmap56:
		value = c.ReadUint(binary.LittleEndian, 7)
	case ZclDataTypeBitmap64:
		value = c.ReadUint(binary.LittleEndian, 8)
	case ZclDataTypeUint8:
		value = c.ReadUint(binary.LittleEndian, 1)
	case ZclDataTypeUint16:
		value = c.ReadUint(binary.LittleEndian, 2)
	case ZclDataTypeUint24:
		value = c.ReadUint(binary.LittleEndian, 3)
	case ZclDataTypeUint32:
		value = c.ReadUint(binary.LittleEndian, 4)
	case ZclDataTypeUint40:
		value = c.ReadUint(binary.LittleEndian, 5)
	case ZclDataTypeUint48:
		value = c.ReadUint(binary.LittleEndian, 6)
	case ZclDataTypeUint56:
		value = c.ReadUint(binary.LittleEndian, 7)
	case ZclDataTypeUint64:
		value = c.ReadUint(binary.LittleEndian, 8)
	case ZclDataTypeInt8:
		value = c.ReadInt(binary.LittleEndian, 1)
	case ZclDataTypeInt16:
		value = c.ReadInt(binary.LittleEndian, 2)
	case ZclDataTypeInt24:
		value = c.ReadInt(binary.LittleEndian, 3)
	case ZclDataTypeInt32:
		value = c.ReadInt(binary.LittleEndian, 4)
	case ZclDataTypeInt40:
		value = c.ReadInt(binary.LittleEndian, 5)
	case ZclDataTypeInt48:
		value = c.ReadInt(binary.LittleEndian, 6)
	case ZclDataTypeInt56:
		value = c.ReadInt(binary.LittleEndian, 7)
	case ZclDataTypeInt64:
		value = c.ReadInt(binary.LittleEndian, 8)
	case ZclDataTypeEnum8:
		value = c.ReadUint(binary.LittleEndian, 1)
	case ZclDataTypeEnum16:
		value = c.ReadUint(binary.LittleEndian, 2)
	case ZclDataTypeSemiPrec:
	case ZclDataTypeSinglePrec:
	case ZclDataTypeDoublePrec:
	case ZclDataTypeOctetStr:
		len, _ := c.ReadByte()
		value, _ = c.ReadString(int(len))
	case ZclDataTypeCharStr:
		len, _ := c.ReadByte()
		value, _ = c.ReadString(int(len))
	case ZclDataTypeLongOctetStr:
		len, _ := c.ReadUint16le()
		value, _ = c.ReadString(int(len))
	case ZclDataTypeLongCharStr:
		len, _ := c.ReadUint16le()
		value, _ = c.ReadString(int(len))
	case ZclDataTypeArray, ZclDataTypeSet, ZclDataTypeBag:
		len, _ := c.ReadUint16le()
		arr := make([]*Attribute, len)
		for i := 0; i < int(len); i++ {
			attribute := &Attribute{}
			attribute.DataType, attribute.Value = readAttribute(c)
			arr[i] = attribute
		}
		value = arr
	case ZclDataTypeStruct:
	case ZclDataTypeTod:
		hours, _ := c.ReadUint8()
		minutes, _ := c.ReadUint8()
		seconds, _ := c.ReadUint8()
		hundredths, _ := c.ReadUint8()
		value = &TimeOfDay{hours, minutes, seconds, hundredths}
	case ZclDataTypeDate:
		year, _ := c.ReadUint8()
		month, _ := c.ReadUint8()
		dayOfMonth, _ := c.ReadUint8()
		dayOfWeek, _ := c.ReadUint8()
		value = &Date{year, month, dayOfMonth, dayOfWeek}
	case ZclDataTypeUtc:
		value, _ = c.ReadUint32le()
	case ZclDataTypeClusterId:
		value, _ = c.ReadUint16le()
	case ZclDataTypeAttrId:
		value, _ = c.ReadUint16le()
	case ZclDataTypeBacOid:
		value, _ = c.ReadUint32le()
	case ZclDataTypeIeeeAddr:
		v, _ := c.ReadUint64le()
		value, _ = binstruct.UintToHexString(v, 8)
	case ZclDataType_128BitSecKey:
		var key [16]byte
		_ = c.ReadBuf(key[:])
		value = key
	case ZclDataTypeUnknown:

	}
	return
}

func flag(boolean bool) uint8 {
	if boolean {
		return 1
	}
	return 0
}
