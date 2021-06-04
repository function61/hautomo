package ezhub

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"math"
	"math/big"
	"time"

	"github.com/function61/gokit/encoding/jsonfile"
	. "github.com/function61/hautomo/pkg/builtin"
	"github.com/function61/hautomo/pkg/changedetector"
	"github.com/function61/hautomo/pkg/ezstack/coordinator"
	"github.com/function61/hautomo/pkg/ezstack/zigbee"
)

type Config struct {
	Coordinator coordinator.Configuration
	HttpAddr    string      `json:"HttpAddr,omitempty"`
	MQTT        *MQTTConfig `json:"MQTT,omitempty"`
}

func (c Config) Valid() error {
	return FirstError(
		c.Coordinator.Valid(),
		c.MQTT.Valid())
}

type MQTTConfig struct {
	Prefix string // e.g. "ezhub" => device states are published to ezhub/<device>
	Addr   string // e.g. "127.0.0.1:1883"
}

func (m MQTTConfig) Valid() error {
	return FirstError(
		UnsetErrorIf(m.Prefix == "", "Prefix"),
		UnsetErrorIf(m.Addr == "", "Addr"))
}

func createStateSnapshotTask(nodeDatabase *nodeDb) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		saveTimer := time.NewTicker(30 * time.Second)

		stateChangeDetector := changedetector.New()

		saveIfChanged := func() error {
			changed, err := stateChangeDetector.WriterChanged(func(sink io.Writer) error {
				return jsonfile.Marshal(sink, nodeDatabase)
			})
			if err != nil {
				return err
			}

			if changed {
				return saveNodeDatabase(nodeDatabase)
			} else {
				return nil
			}
		}

		for {
			select {
			case <-ctx.Done():
				return saveIfChanged()
			case <-saveTimer.C:
				if err := saveIfChanged(); err != nil {
					return err // unable to write to disk is worth a crash
				}
			}
		}
	}
}

func GenerateConfiguration(output io.Writer) error {
	coordinatorIEEEAddress := make([]byte, 8)
	if _, err := rand.Read(coordinatorIEEEAddress); err != nil {
		return err
	}

	networkKey := make([]byte, 16)
	if _, err := rand.Read(networkKey); err != nil {
		return err
	}

	panId, err := rand.Int(rand.Reader, big.NewInt(math.MaxUint16))
	if err != nil {
		return err
	}

	extPanId, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return err
	}

	conf := Config{
		Coordinator: coordinator.Configuration{
			NetworkConfiguration: coordinator.NetworkConfiguration{
				PanId:       zigbee.PANID(panId.Int64()),
				ExtPanId:    zigbee.ExtendedPANID(extPanId.Uint64()),
				IEEEAddress: fmt.Sprintf("0x%x", coordinatorIEEEAddress),
				NetworkKey:  networkKey,
				Channel:     15,
			},
			Serial: &coordinator.Serial{
				Port: "/dev/ttyACM0",
			},
		},
		MQTT: &MQTTConfig{
			Prefix: "ezhub1",
			Addr:   "127.0.0.1:1883",
		},
	}

	return jsonfile.Marshal(output, conf)
}
