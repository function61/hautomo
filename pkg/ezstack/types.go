package ezstack

// TODO: most of these probably belong in a "zigbee" package

import (
	"github.com/function61/hautomo/pkg/ezstack/zcl"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
	"github.com/function61/hautomo/pkg/ezstack/zigbee"
)

// combines device and endpoint, since they both are needed when communicating with a device
type DeviceAndEndpoint struct {
	NetworkAddress string
	EndpointId     zigbee.EndpointId
}

type Model string

// I think the value proposition of "endpoint" is because you simply cannot set OnOff cluster's "on"
// attribute for *the entire Zigbee device* - what if it's a Zigbee power strip? E.g. with endpoints
// the Zigbee device (= power strip) can present its multiple sockets separately
type Endpoint struct {
	Id            zigbee.EndpointId
	ProfileId     uint16
	DeviceId      uint16
	DeviceVersion uint8

	// clusters are groupings of attributes, i.e. a feature. a temperature sensor supports temperature cluster
	// which has many related attributes (measuredValue, minMeasuredValue, maxMeasuredValue). a temperature
	// sensor usually also implements relative humidity cluster..
	// https://twitter.com/joonas_fi/status/1349630750832947201

	InClusterList  []cluster.ClusterId // input, i.e. what endpoint attributes the device can receive from us
	OutClusterList []cluster.ClusterId // output, i.e. what endpoint attributes the device can send to us
}

type DeviceIncomingMessage struct {
	Device          *Device
	IncomingMessage *zcl.ZclIncomingMessage
}

type PowerSource uint8

const (
	Unknown PowerSource = iota
	MainsSinglePhase
	Mains2Phase
	Battery
	DCSource
	EmergencyMainsConstantlyPowered
	EmergencyMainsAndTransfer
)

type Device struct {
	IEEEAddress    string // "MAC address" that never changes. longer form of *NetworkAddress*, but curiously not present in incoming messages, so what is it used for?
	NetworkAddress string // shorter address that is used in Zigbee frames. this changes each time the device leaves and then enters the network
	Manufacturer   string
	ManufacturerId uint16
	Model          Model
	LogicalType    zigbee.LogicalType // coordinator | router | end device
	MainPowered    bool
	PowerSource    PowerSource
	Endpoints      []*Endpoint
}

var powerSourceStrings = map[PowerSource]string{
	Unknown:                         "Unknown",
	MainsSinglePhase:                "MainsSinglePhase",
	Mains2Phase:                     "Mains2Phase",
	Battery:                         "Battery",
	DCSource:                        "DCSource",
	EmergencyMainsConstantlyPowered: "EmergencyMainsConstantlyPowered",
	EmergencyMainsAndTransfer:       "EmergencyMainsAndTransfer",
}

func (ps PowerSource) String() string {
	return powerSourceStrings[ps]
}
