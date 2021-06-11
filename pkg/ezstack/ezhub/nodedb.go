package ezhub

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/function61/gokit/encoding/jsonfile"
	"github.com/function61/gokit/sync/syncutil"
	"github.com/function61/hautomo/pkg/ezstack"
	"github.com/function61/hautomo/pkg/ezstack/ezhub/hubtypes"
	"github.com/function61/hautomo/pkg/ezstack/zigbee"
)

const (
	nodeDbFilename = "ezhub-state.json"
)

type nodeDb struct {
	Devices []*hubtypes.Device `json:"devices"`
	mu      sync.Mutex
}

func loadNodeDatabase() (*nodeDb, error) {
	db := &nodeDb{}
	if err := jsonfile.ReadDisallowUnknownFields(nodeDbFilename, db); err != nil {
		return nil, err
	}

	return db, nil
}

func saveNodeDatabase(db *nodeDb) error {
	return jsonfile.Write(nodeDbFilename, db)
}

var _ ezstack.NodeDatabase = (*nodeDb)(nil)

func (d *nodeDb) InsertDevice(device *ezstack.Device) error {
	// TODO: locking is borked, but better than nothing

	if _, found := d.GetDevice(device.IEEEAddress); found {
		// likely re-joined network, so just update record in DB
		//
		// NO-OP: just assuming that user modified record from GetDevice() which already shares
		//        the pointer value
		return nil
	}

	if _, found := d.GetDeviceByNetworkAddress(device.NetworkAddress); found {
		return fmt.Errorf("already exists by NetworkAddress: %s", device.NetworkAddress)
	}

	defer lockAndUnlock(&d.mu)()

	attrsPerEndpoint := map[zigbee.EndpointId]*hubtypes.Attributes{}
	for _, endpointSpec := range device.Endpoints { // each endpoint gets its own attributes set
		attrsPerEndpoint[endpointSpec.Id] = hubtypes.NewAttributes()
	}

	d.Devices = append(d.Devices, &hubtypes.Device{
		FriendlyName: device.IEEEAddress.HexPrefixedString(), // start with something for friendly name
		ZigbeeDevice: device,

		State: &hubtypes.DeviceState{
			LinkQuality: &hubtypes.AttrInt{Value: 0, LastReport: time.Now().UTC()}, // we just had a chat with the device

			EndpointAttrs: attrsPerEndpoint,
		},
	})

	return saveNodeDatabase(d)
}

func (d *nodeDb) GetDevice(ieeeAddress zigbee.IEEEAddress) (*ezstack.Device, bool) {
	// lock implemented in subcall

	if wdev := d.GetWrappedDevice(ieeeAddress); wdev != nil {
		return wdev.ZigbeeDevice, true
	}

	return nil, false
}

func (d *nodeDb) GetWrappedDevice(ieeeAddress zigbee.IEEEAddress) *hubtypes.Device {
	defer lockAndUnlock(&d.mu)()

	for _, dev := range d.Devices {
		if dev.ZigbeeDevice.IEEEAddress == ieeeAddress {
			return dev
		}
	}

	return nil
}

func (d *nodeDb) GetDeviceByNetworkAddress(nwkAddress string) (*ezstack.Device, bool) {
	defer lockAndUnlock(&d.mu)()

	for _, dev := range d.Devices {
		if dev.ZigbeeDevice.NetworkAddress == nwkAddress {
			return dev.ZigbeeDevice, true
		}
	}

	return nil, false
}

func (d *nodeDb) RemoveDevice(ieeeAddress zigbee.IEEEAddress) error {
	defer lockAndUnlock(&d.mu)()

	for idx, dev := range d.Devices {
		if dev.ZigbeeDevice.IEEEAddress == ieeeAddress {
			d.Devices = append(d.Devices[:idx], d.Devices[idx+1:]...)
			return nil
		}
	}

	return fmt.Errorf("not found by: %s", ieeeAddress)
}

func (d *nodeDb) withLock(do func() error) error {
	defer lockAndUnlock(&d.mu)()

	return do()
}

func loadNodeDatabaseOrInitIfNotFound() (*nodeDb, error) {
	db, err := loadNodeDatabase()
	if err != nil {
		if os.IsNotExist(err) {
			if err := saveNodeDatabase(&nodeDb{Devices: []*hubtypes.Device{}}); err != nil {
				return nil, err
			}

			return loadNodeDatabase() // try "one last time"
		} else { // some other error
			return nil, err
		}
	}

	return db, nil
}

var lockAndUnlock = syncutil.LockAndUnlock
