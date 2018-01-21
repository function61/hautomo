package types

type LightRequest struct {
	Red           uint8
	Green         uint8
	Blue          uint8
	BluetoothAddr string
}

func LightRequestNew(bluetoothAddr string, r uint8, g uint8, b uint8) LightRequest {
	return LightRequest{r, g, b, bluetoothAddr}
}

func LightRequestOn(bluetoothAddr string) LightRequest {
	return LightRequestNew(bluetoothAddr, 255, 255, 255)
}

func LightRequestOff(bluetoothAddr string) LightRequest {
	return LightRequestNew(bluetoothAddr, 0, 0, 0)
}

func (l *LightRequest) IsOff() bool {
	return (l.Red + l.Green + l.Blue) == 0
}

func (l *LightRequest) IsOn() bool {
	return uint16(l.Red)+uint16(l.Green)+uint16(l.Blue) == (255 * 3)
}
