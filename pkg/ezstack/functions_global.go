package ezstack

// global commands are commands that are present in every cluster, like *ReadAttributes* and
// *WriteAttributes*.

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/function61/hautomo/pkg/ezstack/binstruct"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
	"github.com/function61/hautomo/pkg/ezstack/zcl/frame"
	"github.com/function61/hautomo/pkg/ezstack/znp"
)

// ZCL spec section 2.5.1
func (s *Stack) ReadAttributes(
	nwkAddress string,
	clusterId cluster.ClusterId,
	attributeIds []cluster.AttributeId,
) (*cluster.ReadAttributesResponse, error) {
	response, err := s.globalCommand(nwkAddress, clusterId, 0x00, &cluster.ReadAttributesCommand{castAttributeIds(attributeIds)})
	if err != nil {
		return nil, err
	}

	return response.(*cluster.ReadAttributesResponse), nil
}

// ZCL spec section 2.5.3
func (s *Stack) WriteAttributes(nwkAddress string, clusterId cluster.ClusterId, writeAttributeRecords []*cluster.WriteAttributeRecord) (*cluster.WriteAttributesResponse, error) {
	response, err := s.globalCommand(nwkAddress, clusterId, 0x02, &cluster.WriteAttributesCommand{writeAttributeRecords})
	if err != nil {
		return nil, err
	}

	return response.(*cluster.WriteAttributesResponse), nil
}

// ZCL spec section 2.5.7
func (s *Stack) ConfigureReporting(nwkAddress string, clusterId cluster.ClusterId, configs ...*cluster.AttributeReportingConfigurationRecord) error {
	response, err := s.globalCommand(nwkAddress, clusterId, 0x06, &cluster.ConfigureReportingCommand{configs})
	if err != nil {
		return err
	}

	// TODO: validate status
	_ = response.(*cluster.ConfigureReportingResponse)

	return nil
}

func (s *Stack) globalCommand(nwkAddress string, clusterId cluster.ClusterId, commandId uint8, command interface{}) (interface{}, error) {
	options := &znp.AfDataRequestOptions{}
	frm, err := frame.New().
		DisableDefaultResponse(true).
		FrameType(frame.FrameTypeGlobal).
		Direction(frame.DirectionClientServer).
		CommandId(commandId).
		Command(command).
		Build()

	if err != nil {
		return nil, err
	}

	response, err := s.coordinator.DataRequest(nwkAddress, 255, 1, uint16(clusterId), options, 15, binstruct.Encode(frm))
	if err == nil {
		zclIncomingMessage, err := s.zcl.ToZclIncomingMessage(response)
		if err == nil {
			return zclIncomingMessage.Data.Command, nil
		} else {
			logl.Error.Printf("Unsupported data response message:\n%s\n", func() string { return spew.Sdump(response) })
		}

	}
	return nil, err
}

func castAttributeIds(items []cluster.AttributeId) []uint16 {
	cast := []uint16{}
	for _, item := range items {
		cast = append(cast, uint16(item))
	}
	return cast
}
