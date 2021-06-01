package cluster

import "time"

type Definition struct {
	name       string
	attributes map[AttributeId]*AttributeDescriptor
}

func (d *Definition) Name() string {
	return d.name
}

func (d *Definition) Attribute(id AttributeId) *AttributeDescriptor {
	return d.attributes[id]
}

type Access uint8

const (
	Read       Access = 0x01
	Write      Access = 0x02
	Reportable Access = 0x04
	Scene      Access = 0x08
)

type ClusterId uint16

type AttributeDescriptor struct {
	Name   string
	Type   ZclDataType
	Access Access
}

type AttributeId uint16

// these are used from elsewhere, so prevent magic constants
const (
	AttrBasicManufacturerName AttributeId = 4
	AttrBasicModelId          AttributeId = 5
	AttrBasicPowerSource      AttributeId = 7
)

// Zigbee transition times are in units of 100 milliseconds
func TransitionTimeFrom(duration time.Duration) uint16 {
	return uint16(duration.Milliseconds() / 100)
}
