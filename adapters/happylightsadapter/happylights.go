package happylightsadapter

import (
	"github.com/function61/home-automation-hub/hapitypes"
	"github.com/function61/home-automation-hub/libraries/happylights/client"
	"github.com/function61/home-automation-hub/libraries/happylights/types"
	"log"
)

func New(adapter *hapitypes.Adapter, config hapitypes.AdapterConfig) *hapitypes.Adapter {
	go func() {
		log.Println("HappyLightsAdapter: started")

		for {
			select {
			case powerMsg := <-adapter.PowerMsg:
				bluetoothAddr := powerMsg.DeviceId

				var req types.LightRequest

				if powerMsg.On {
					req = types.LightRequestOn(bluetoothAddr)
				} else {
					req = types.LightRequestOff(bluetoothAddr)
				}

				if err := client.SendRequest(config.HappyLightsAddr, req); err != nil {
					log.Printf("HappyLightsAdapter: error %s", err.Error())
				}
			case colorMsg := <-adapter.ColorMsg:
				bluetoothAddr := colorMsg.DeviceId

				// convert to happylights request
				hlreq := types.LightRequestColor(
					bluetoothAddr,
					colorMsg.Color.Red,
					colorMsg.Color.Green,
					colorMsg.Color.Blue)

				if err := client.SendRequest(config.HappyLightsAddr, hlreq); err != nil {
					log.Printf("HappyLightsAdapter: error %s", err.Error())
				}
			}
		}
	}()

	return adapter
}
