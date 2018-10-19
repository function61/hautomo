package irsimulatoradapter

import (
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/hapitypes"
	"github.com/function61/home-automation-hub/pkg/signalfabric"
	"time"
)

func StartSensor(adapter *hapitypes.Adapter, config hapitypes.AdapterConfig, fabric *signalfabric.Fabric, stop *stopper.Stopper) {
	go func() {
		defer stop.Done()

		for {
			select {
			case <-time.After(5 * time.Second):
				fabric.InfraredEvent <- hapitypes.NewInfraredEvent("simulated_remote", config.IrSimulatorKey)
			case <-stop.Signal:
				return
			}
		}
	}()
}
