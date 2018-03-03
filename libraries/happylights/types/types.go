package types

type LightRequest struct {
	Red           uint8
	Green         uint8
	Blue          uint8
	BluetoothAddr string
}

func lightRequestNew(bluetoothAddr string, r uint8, g uint8, b uint8) LightRequest {
	return LightRequest{r, g, b, bluetoothAddr}
}

func LightRequestColor(bluetoothAddr string, r uint8, g uint8, b uint8) LightRequest {
	req := lightRequestNew(bluetoothAddr, r, g, b)

	// full white? this would be understood as "turn on", using previous color.
	// use almost full white to tell that we actually mean white
	if req.IsOn() {
		return lightRequestNew(bluetoothAddr, 255, 255, 254)
	}

	return req
}

func LightRequestOn(bluetoothAddr string) LightRequest {
	return lightRequestNew(bluetoothAddr, 255, 255, 255)
}

func LightRequestOff(bluetoothAddr string) LightRequest {
	return lightRequestNew(bluetoothAddr, 0, 0, 0)
}

func (l *LightRequest) IsOff() bool {
	return (l.Red + l.Green + l.Blue) == 0
}

func (l *LightRequest) IsOn() bool {
	return uint16(l.Red)+uint16(l.Green)+uint16(l.Blue) == (255 * 3)
}
