package ezstack

// Registration is somewhat the most complex operation of this package, so it deserves its own file

import (
	"fmt"

	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
	"github.com/function61/hautomo/pkg/ezstack/znp"
)

// queries "device metadata" such as description, endpoints, supported clusters
func (s *Stack) interrogateDevice(announcedDevice *znp.ZdoEndDeviceAnnceInd) (*Device, error) {
	ieeeAddress := announcedDevice.IEEEAddr
	nwkAddress := announcedDevice.NwkAddr

	deviceDetails, err := s.ReadAttributes(nwkAddress, cluster.IdGenBasic, []cluster.AttributeId{
		cluster.AttrBasicManufacturerName,
		cluster.AttrBasicModelId,
		cluster.AttrBasicPowerSource,
	})
	if err != nil {
		return nil, fmt.Errorf("querying basic metadata: %w", err)
	}

	findAttr := func(id cluster.AttributeId) *cluster.Attribute { // ugh
		for _, attr := range deviceDetails.ReadAttributeStatuses {
			if attr.AttributeID == uint16(id) {
				return attr.Attribute
			}
		}

		return nil
	}

	manufacturer := func() string {
		if val, ok := findAttr(cluster.AttrBasicManufacturerName).Value.(string); ok {
			return val
		}
		return ""
	}()
	deviceModel := func() string {
		if val, ok := findAttr(cluster.AttrBasicModelId).Value.(string); ok {
			return val
		}
		return ""
	}()
	powerSource := func() PowerSource {
		if val, ok := findAttr(cluster.AttrBasicPowerSource).Value.(uint64); ok {
			return PowerSource(val)
		}
		return PowerSource(0) // this is what unset branch did in previous implementation anyway
	}()

	logl.Debug.Printf("Querying node description: [%s]", ieeeAddress)

	nodeDescription, err := s.coordinator.NodeDescription(nwkAddress)
	if err != nil {
		return nil, fmt.Errorf("querying node description: %w", err)
	}

	logl.Debug.Printf("Querying active endpoints: [%s]", ieeeAddress)

	activeEndpoints, err := s.coordinator.ActiveEndpoints(nwkAddress)
	if err != nil {
		return nil, fmt.Errorf("querying active endpoints: %w", err)
	}

	endpoints := []*Endpoint{}
	for _, endpointNo := range activeEndpoints.ActiveEPList {
		logl.Debug.Printf("Request endpoint description: [%s], ep: [%d]", ieeeAddress, endpointNo)

		endpointDescr, err := s.coordinator.SimpleDescription(nwkAddress, endpointNo)
		if err != nil {
			return nil, fmt.Errorf("query endpoint %d description: %w", endpointNo, err)
		}

		endpoints = append(endpoints, &Endpoint{
			Id:             endpointDescr.Endpoint,
			ProfileId:      endpointDescr.ProfileID,
			DeviceId:       endpointDescr.DeviceID,
			DeviceVersion:  endpointDescr.DeviceVersion,
			InClusterList:  castClusterIds(endpointDescr.InClusterList),
			OutClusterList: castClusterIds(endpointDescr.OutClusterList),
		})
	}

	return &Device{
		IEEEAddress:    ieeeAddress,
		NetworkAddress: nwkAddress,
		MainPowered:    announcedDevice.Capabilities.MainPowered > 0,
		Manufacturer:   manufacturer,
		Model:          Model(deviceModel),
		PowerSource:    powerSource,
		LogicalType:    nodeDescription.LogicalType,
		ManufacturerId: nodeDescription.ManufacturerCode,
		Endpoints:      endpoints,
	}, nil
}
