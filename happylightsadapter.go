package main

import (
	"./happylights/client"
	"./happylights/types"
	"log"
)

func NewHappylightsAdapter(id string, serverAddr string) *Adapter {
	adapter := NewAdapter(id)

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

				if err := client.SendRequest(serverAddr, req); err != nil {
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

				if err := client.SendRequest(serverAddr, hlreq); err != nil {
					log.Printf("HappyLightsAdapter: error %s", err.Error())
				}
			}
		}
	}()

	return adapter
}
