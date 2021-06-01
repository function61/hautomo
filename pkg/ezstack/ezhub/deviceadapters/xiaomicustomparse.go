package deviceadapters

import (
	"github.com/function61/hautomo/pkg/ezstack/ezhub/hubtypes"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
	"github.com/shimmeringbee/bytecodec"
	"github.com/shimmeringbee/zcl"
)

type Attribute struct {
	Id        uint8
	Attribute *zcl.AttributeDataTypeValue
}

type AttributeList []Attribute

func (a AttributeList) Find(id uint8) *zcl.AttributeDataTypeValue {
	for _, item := range a {
		if item.Id == id {
			return item.Attribute
		}
	}

	return nil
}

func ParseAttributeList(xiaomiBytes []byte) (AttributeList, error) {
	var xal AttributeList

	return xal, bytecodec.Unmarshal(xiaomiBytes, &xal)
}

// Known attribute numbers:
// 11 = illuminance
// 3 = temperature
// 100 = contact
// 102 = pressure
const (
	aqaraCustomAttrNrBatteryVoltage = 1
)

// parses custom Xiaomi binary list of attributes into battery voltage etc.
var aqaraVoltageEtc = attributeParser("genBasic.unknown(65281)", aqaraCustomAttributesParse)

func aqaraCustomAttributesParse(attr *cluster.Attribute, actx *hubtypes.AttrsCtx) error {
	attrList, err := ParseAttributeList([]byte(attr.Value.(string)))
	if err != nil {
		return err
	}

	for _, attr := range attrList {
		switch attr.Id {
		case aqaraCustomAttrNrBatteryVoltage:
			actx.Attrs.BatteryVoltage = actx.Float(float64(attr.Attribute.Value.(uint64)) / 1000)
			// other values observed: unknown(3)=35 unknown(4)=5032 unknown(5)=72 unknown(6)=0 unknown(10)=0
		}
	}

	return nil
}
