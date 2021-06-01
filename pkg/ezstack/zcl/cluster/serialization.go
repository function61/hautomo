package cluster

import (
	"fmt"
)

// takes in way too magical interface{} which is one of Go's built-in representation types that were
// unmarshaled from []byte based on ZclDataType
//
// TODO: move this somewhere
// TODO: test this?
// TODO: document why this works
// TODO: refactor code so eventually this is not required
func SerializeMagicValue(dataType ZclDataType, val interface{}) string {
	// interface{} by default

	switch dataType {
	case ZclDataTypeNoData:
		return "nil(NoData)"
	case ZclDataTypeData8:
		return fmt.Sprintf("%x", val.([1]byte))
	case ZclDataTypeData16:
		return fmt.Sprintf("%x", val.([2]byte))
	case ZclDataTypeData24:
		return fmt.Sprintf("%x", val.([3]byte))
	case ZclDataTypeData32:
		return fmt.Sprintf("%x", val.([4]byte))
	case ZclDataTypeData40:
		return fmt.Sprintf("%x", val.([5]byte))
	case ZclDataTypeData48:
		return fmt.Sprintf("%x", val.([6]byte))
	case ZclDataTypeData56:
		return fmt.Sprintf("%x", val.([7]byte))
	case ZclDataTypeData64:
		return fmt.Sprintf("%x", val.([8]byte))
	case ZclDataTypeBoolean:
		return fmt.Sprintf("%s", func() string {
			if val.(bool) {
				return "true"
			} else {
				return "false"
			}
		}())
	case ZclDataTypeBitmap8:
		return fmt.Sprintf("Bitmap8(%#b)", val.(uint64))
	case ZclDataTypeBitmap16:
		return fmt.Sprintf("Bitmap16(%#b)", val.(uint64))
	case ZclDataTypeBitmap24:
		return fmt.Sprintf("Bitmap24(%#b)", val.(uint64))
	case ZclDataTypeBitmap32:
		return fmt.Sprintf("Bitmap32(%#b)", val.(uint64))
	case ZclDataTypeBitmap40:
		return fmt.Sprintf("Bitmap40(%#b)", val.(uint64))
	case ZclDataTypeBitmap48:
		return fmt.Sprintf("Bitmap48(%#b)", val.(uint64))
	case ZclDataTypeBitmap56:
		return fmt.Sprintf("Bitmap56(%#b)", val.(uint64))
	case ZclDataTypeBitmap64:
		return fmt.Sprintf("Bitmap64(%#b)", val.(uint64))
	case ZclDataTypeUint8:
		return fmt.Sprintf("Uint8(%d)", val.(uint64))
	case ZclDataTypeUint16:
		return fmt.Sprintf("Uint16(%d)", val.(uint64))
	case ZclDataTypeUint24:
		return fmt.Sprintf("Uint24(%d)", val.(uint64))
	case ZclDataTypeUint32:
		return fmt.Sprintf("Uint32(%d)", val.(uint64))
	case ZclDataTypeUint40:
		return fmt.Sprintf("Uint40(%d)", val.(uint64))
	case ZclDataTypeUint48:
		return fmt.Sprintf("Uint48(%d)", val.(uint64))
	case ZclDataTypeUint56:
		return fmt.Sprintf("Uint56(%d)", val.(uint64))
	case ZclDataTypeUint64:
		return fmt.Sprintf("Uint64(%d)", val.(uint64))
	case ZclDataTypeInt8:
		return fmt.Sprintf("Int8(%d)", val.(int64))
	case ZclDataTypeInt16:
		return fmt.Sprintf("Int16(%d)", val.(int64))
	case ZclDataTypeInt24:
		return fmt.Sprintf("Int24(%d)", val.(int64))
	case ZclDataTypeInt32:
		return fmt.Sprintf("Int32(%d)", val.(int64))
	case ZclDataTypeInt40:
		return fmt.Sprintf("Int40(%d)", val.(int64))
	case ZclDataTypeInt48:
		return fmt.Sprintf("Int48(%d)", val.(int64))
	case ZclDataTypeInt56:
		return fmt.Sprintf("Int56(%d)", val.(int64))
	case ZclDataTypeInt64:
		return fmt.Sprintf("Int64(%d)", val.(int64))
	case ZclDataTypeEnum8:
		return fmt.Sprintf("Enum8(%d)", val.(uint64))
	case ZclDataTypeEnum16:
		return fmt.Sprintf("Enum16(%d)", val.(uint64))
	case ZclDataTypeSemiPrec:
		return fmt.Sprintf("SemiPrec(%s)", val.(string))
	case ZclDataTypeSinglePrec:
		return fmt.Sprintf("SinglePrec(%s)", val.(string))
	case ZclDataTypeDoublePrec:
		return fmt.Sprintf("DoublePrec(%s)", val.(string))
	case ZclDataTypeOctetStr:
		return fmt.Sprintf("OctetStr(%s)", val.(string))
	case ZclDataTypeCharStr:
		return fmt.Sprintf("CharStr(%s)", val.(string))
	case ZclDataTypeLongOctetStr:
		return fmt.Sprintf("LongOctetStr(%s)", val.(string))
	case ZclDataTypeLongCharStr:
		return fmt.Sprintf("LongCharStr(%s)", val.(string))
	case ZclDataTypeArray, ZclDataTypeSet, ZclDataTypeBag:
		items := val.([]*Attribute)
		serialized := []string{}
		for _, item := range items {
			serialized = append(serialized, SerializeMagicValue(item.DataType, item.Value))
		}

		return fmt.Sprintf("Array|TypeSet|Bag(%v)", serialized)
	case ZclDataTypeStruct:
		return fmt.Sprintf("%v", val)
	case ZclDataTypeTod:
		return fmt.Sprintf("TimeOfDay(%v)", val.(*TimeOfDay))
	case ZclDataTypeDate:
		return fmt.Sprintf("Date(%v)", val.(*Date))
	case ZclDataTypeUtc:
		return fmt.Sprintf("Utc(%d)", val.(uint32))
	case ZclDataTypeClusterId:
		return fmt.Sprintf("ClusterId(%d)", val.(uint16))
	case ZclDataTypeAttrId:
		return fmt.Sprintf("AttrId(%d)", val.(uint16))
	case ZclDataTypeBacOid:
		return fmt.Sprintf("BacOid(%d)", val.(uint32))
	case ZclDataTypeIeeeAddr:
		return fmt.Sprintf("IeeeAddr(%s)", val.(string))
	case ZclDataType_128BitSecKey:
		return fmt.Sprintf("128BitSecKey(%x)", val.([16]byte))
	case ZclDataTypeUnknown:
		return "interface{}"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", dataType)
	}
}
