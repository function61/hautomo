package ezstack

import (
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
	"github.com/function61/hautomo/pkg/ezstack/zigbee"
)

/* TODO: has source NwkAddr confused with destination NwkAddr
// "Zigbee has support for binding which makes it possible that devices can directly control each
// other without the intervention of Zigbee2MQTT or any home automation software."
// https://www.zigbee2mqtt.io/information/binding.html
func (s *Stack) Bind(
	sourceAddress zigbee.IEEEAddress,
	sourceEndpoint zigbee.EndpointId,
	clusterId cluster.ClusterId,
	destinationIeeeAddress zigbee.IEEEAddress,
	destinationEndpoint zigbee.EndpointId,
) (*znp.ZdoBindRsp, error) {
	dev, found := s.db.GetDevice(sourceAddress)
	if !found {
		return nil, fmt.Errorf("device not found: %s", sourceAddress)
	}

	return s.coordinator.Bind(
		dev.NetworkAddress,
		sourceAddress,
		sourceEndpoint,
		clusterId,
		destinationIeeeAddress,
		destinationEndpoint)
}
*/

func (s *Stack) BindToCoordinator(
	sourceAddress zigbee.IEEEAddress,
	sourceEndpoint zigbee.EndpointId,
	clusterId cluster.ClusterId,
	coordinatorEndpoint zigbee.EndpointId,
) error {
	resp, err := s.coordinator.Bind(
		zigbee.CoordinatorNwkAddr,
		sourceAddress,
		sourceEndpoint,
		clusterId,
		s.coordinator.NetworkConf().IEEEAddress,
		coordinatorEndpoint)
	if err != nil {
		return err
	}

	return resp.Status.Error()
}

/* TODO: has source NwkAddr confused with destination NwkAddr
// undoes the effect of Bind()
func (s *Stack) Unbind(
	sourceAddress zigbee.IEEEAddress,
	sourceEndpoint zigbee.EndpointId,
	clusterId cluster.ClusterId,
	destinationIeeeAddress zigbee.IEEEAddress,
	destinationEndpoint zigbee.EndpointId,
) (*znp.ZdoUnbindRsp, error) {
	dev, found := s.db.GetDevice(sourceAddress)
	if !found {
		return nil, fmt.Errorf("device not found: %s", sourceAddress)
	}

	return s.coordinator.Unbind(
		dev.NetworkAddress,
		sourceAddress,
		sourceEndpoint,
		clusterId,
		destinationIeeeAddress,
		destinationEndpoint)
}
*/
