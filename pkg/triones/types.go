package triones

type RequestKind int

const (
	RequestKindOn RequestKind = iota
	RequestKindOff
	RequestKindRGB
	RequestKindWhite
)

type Request struct {
	Kind          RequestKind
	BluetoothAddr string
	RgbOpts       *RgbOpts
	WhiteOpts     *WhiteOpts // use only if the strip has separate white channel (RGBW strip)
}

type RgbOpts struct {
	Red   uint8
	Green uint8
	Blue  uint8
}

type WhiteOpts struct {
	Brightness uint8
}

func RequestOn(bluetoothAddr string) Request {
	return Request{
		Kind:          RequestKindOn,
		BluetoothAddr: bluetoothAddr,
	}
}

func RequestOff(bluetoothAddr string) Request {
	return Request{
		Kind:          RequestKindOff,
		BluetoothAddr: bluetoothAddr,
	}
}

func RequestRGB(bluetoothAddr string, r, g, b uint8) Request {
	return Request{
		Kind:          RequestKindRGB,
		BluetoothAddr: bluetoothAddr,
		RgbOpts: &RgbOpts{
			Red:   r,
			Green: g,
			Blue:  b,
		},
	}
}

func RequestWhite(bluetoothAddr string, brightness uint8) Request {
	return Request{
		Kind:          RequestKindWhite,
		BluetoothAddr: bluetoothAddr,
		WhiteOpts: &WhiteOpts{
			Brightness: brightness,
		},
	}
}
