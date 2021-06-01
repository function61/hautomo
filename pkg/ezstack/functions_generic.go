package ezstack

import (
	"github.com/function61/hautomo/pkg/ezstack/zigbee"
	"github.com/function61/hautomo/pkg/ezstack/znp"
)

// "Zigbee has support for binding which makes it possible that devices can directly control each
// other without the intervention of Zigbee2MQTT or any home automation software."
// https://www.zigbee2mqtt.io/information/binding.html
func (s *Stack) Bind(sourceAddress string, sourceIeeeAddress string, sourceEndpoint zigbee.EndpointId, clusterId uint16, destinationIeeeAddress string, destinationEndpoint zigbee.EndpointId) (*znp.ZdoBindRsp, error) {
	return s.coordinator.Bind(sourceAddress, sourceIeeeAddress, sourceEndpoint, clusterId, znp.AddrModeAddr64Bit, destinationIeeeAddress, destinationEndpoint)
}

func (s *Stack) Unbind(sourceAddress string, sourceIeeeAddress string, sourceEndpoint zigbee.EndpointId, clusterId uint16, destinationIeeeAddress string, destinationEndpoint zigbee.EndpointId) (*znp.ZdoUnbindRsp, error) {
	return s.coordinator.Unbind(sourceAddress, sourceIeeeAddress, sourceEndpoint, clusterId, znp.AddrModeAddr64Bit, destinationIeeeAddress, destinationEndpoint)
}
