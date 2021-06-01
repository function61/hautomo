// Zigbee protocol
package zigbee

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type LogicalType uint8

const (
	LogicalTypeCoordinator LogicalType = 0
	LogicalTypeRouter      LogicalType = 1
	LogicalTypeEndDevice   LogicalType = 2
	LogicalTypeUnknown     LogicalType = 0xff
)

func (l LogicalType) String() string {
	switch l {
	case LogicalTypeCoordinator:
		return "Coordinator"
	case LogicalTypeRouter:
		return "Router"
	case LogicalTypeEndDevice:
		return "EndDevice"
	case LogicalTypeUnknown:
		return "Unknown"
	default:
		panic(fmt.Errorf("unknown LogicalType: %d", l))
	}
}

type NetworkKey [16]byte

type PANID uint16

type ExtendedPANID uint64

func (e ExtendedPANID) MarshalJSON() ([]byte, error) {
	numBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(numBytes, uint64(e))
	return json.Marshal(hex.EncodeToString(numBytes))
}

func (e *ExtendedPANID) UnmarshalJSON(b []byte) error {
	numHex := ""
	if err := json.Unmarshal(b, &numHex); err != nil {
		return err
	}
	numBytes, err := hex.DecodeString(numHex)
	if err != nil {
		return err
	}
	*e = ExtendedPANID(binary.LittleEndian.Uint64(numBytes))
	return nil
}

type EndpointId uint8

type ProfileID uint16

const (
	ProfileIndustrialPlantMonitoring    ProfileID = 0x0101
	ProfileHomeAutomation               ProfileID = 0x0104
	ProfileCommercialBuildingAutomation ProfileID = 0x0105
	ProfileTelecomApplications          ProfileID = 0x0107
	ProfilePersonalHomeAndHospitalCare  ProfileID = 0x0108
	ProfileAdvancedMeteringInitiative   ProfileID = 0x0109
)
