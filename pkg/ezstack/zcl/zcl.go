package zcl

import (
	"fmt"
	"reflect"

	"github.com/function61/hautomo/pkg/ezstack/binstruct"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
	"github.com/function61/hautomo/pkg/ezstack/zcl/frame"
	"github.com/function61/hautomo/pkg/ezstack/zigbee"
	"github.com/function61/hautomo/pkg/ezstack/znp"
)

type CommandExtractor func(commandDescriptors map[uint8]*cluster.CommandDescriptor) (uint8, *cluster.CommandDescriptor, error)

type ClusterQuery func(c map[cluster.ClusterId]*cluster.Cluster) (cluster.ClusterId, *cluster.Cluster, error)

type CommandQuery func(c *cluster.Cluster) (uint8, *cluster.CommandDescriptor, error)

type ZclFrameControl struct {
	FrameType              frame.FrameType
	ManufacturerSpecific   bool
	Direction              frame.Direction
	DisableDefaultResponse bool
}

type ZclFrame struct {
	FrameControl              *ZclFrameControl
	ManufacturerCode          uint16
	TransactionSequenceNumber uint8
	CommandIdentifier         uint8
	CommandName               string // TODO: make Command an interface that has Name()?
	Command                   interface{}
}

// mapped almost 1:1 from znp.AfIncomingMessage, but Data is decoded
type ZclIncomingMessage struct {
	// TODO: instead make most of the shared attributes composed from AfIncomingMessage?
	GroupID              uint16
	ClusterID            cluster.ClusterId
	SrcAddr              string
	SrcEndpoint          zigbee.EndpointId
	DstEndpoint          zigbee.EndpointId
	WasBroadcast         bool
	LinkQuality          uint8
	SecurityUse          bool
	Timestamp            uint32
	TransactionSeqNumber uint8
	Data                 *ZclFrame
}

// the default library. not sure if we ever need additional ones?
var Library = &Zcl{cluster.NewClusterLibrary()}

type Zcl struct {
	library *cluster.ClusterLibrary
}

// returns ~ 1:1 AfIncomingMessage but its Data is parsed as specified by (cluster, command) pair
func (z *Zcl) ToZclIncomingMessage(m *znp.AfIncomingMessage) (*ZclIncomingMessage, error) {
	dataFrame, err := frame.Decode(m.Data)
	if err != nil {
		return nil, err
	}

	// this seems to deserialize generic data into specialized structures with help of command
	// descriptors from cluster library
	dataFrameParsed, name, err := z.inboundFrameToZclCommand(m.ClusterID, dataFrame)
	if err != nil {
		return nil, err
	}

	frameControl := dataFrame.FrameControl // shorthand

	// seems to make about 1:1 copy
	return &ZclIncomingMessage{
		GroupID:              m.GroupID,
		ClusterID:            cluster.ClusterId(m.ClusterID),
		SrcAddr:              m.SrcAddr,
		SrcEndpoint:          m.SrcEndpoint,
		DstEndpoint:          m.DstEndpoint,
		WasBroadcast:         m.WasBroadcast > 0,
		LinkQuality:          m.LinkQuality,
		SecurityUse:          m.SecurityUse > 0,
		Timestamp:            m.Timestamp,
		TransactionSeqNumber: m.TransSeqNumber,
		Data: &ZclFrame{
			FrameControl: &ZclFrameControl{
				FrameType:              frameControl.FrameType,
				ManufacturerSpecific:   frameControl.ManufacturerSpecific > 0,
				Direction:              frameControl.Direction,
				DisableDefaultResponse: frameControl.DisableDefaultResponse > 0,
			},
			ManufacturerCode:          dataFrame.ManufacturerCode,
			TransactionSequenceNumber: dataFrame.TransactionSequenceNumber,
			CommandIdentifier:         dataFrame.CommandIdentifier,
			CommandName:               name,
			Command:                   dataFrameParsed,
		},
	}, nil
}

func (z *Zcl) inboundFrameToZclCommand(clusterId uint16, f *frame.Frame) (interface{}, string, error) {
	var cd *cluster.CommandDescriptor
	var ok bool
	switch f.FrameControl.FrameType {
	case frame.FrameTypeGlobal:
		if cd, ok = z.library.Global()[f.CommandIdentifier]; !ok {
			return nil, "", fmt.Errorf("unsupported global cmd identifier %d", f.CommandIdentifier)
		}
		copy := reflectionCopy(cd.Command) // as not to mutate the one in library

		// somehow this marshals binary data into Go structs. an example being cluster.ReportAttributesCommand
		return copy, cd.Name, binstruct.Decode(f.Payload, copy)
	case frame.FrameTypeLocal:
		c, found := z.library.Clusters()[cluster.ClusterId(clusterId)]
		if !found {
			return nil, "", fmt.Errorf("unknown cluster %d for local cmd", clusterId)
		}
		var commandDescriptors map[uint8]*cluster.CommandDescriptor
		dirDescr := ""
		switch f.FrameControl.Direction {
		case frame.DirectionClientServer:
			commandDescriptors = c.CommandDescriptors.Received
			dirDescr = "received"
		case frame.DirectionServerClient:
			commandDescriptors = c.CommandDescriptors.Generated
			dirDescr = "generated"
		default:
			return nil, "", fmt.Errorf("unrecognized direction: %d", f.FrameControl.Direction)
		}

		if cd, ok = commandDescriptors[f.CommandIdentifier]; !ok {
			return nil, "", fmt.Errorf(
				"cluster %d doesn't support local %s cmd %d (payload %d byte(s))",
				clusterId,
				dirDescr,
				f.CommandIdentifier,
				len(f.Payload))
		}
		copy := reflectionCopy(cd.Command) // as not to mutate the one in library

		// somehow this marshals binary data into Go structs. an example being cluster.ReportAttributesCommand
		return copy, cd.Name, binstruct.Decode(f.Payload, copy)
	default:
		return nil, "", fmt.Errorf("unknown frame type: %d", f.FrameControl.FrameType)
	}
}

func reflectionCopy(n interface{}) interface{} {
	v := reflect.ValueOf(n)
	switch v.Kind() {
	case reflect.Struct:
		copy := reflect.New(v.Type()).Elem()
		return copy.Interface()
	case reflect.Ptr:
		e := v.Elem()
		copy := reflect.New(e.Type())
		return copy.Interface()
	default:
		panic(fmt.Sprintf("reflectionCopy: unsupported value: %#v", n))
	}
}
