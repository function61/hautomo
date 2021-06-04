package hubtypes

import (
	"time"

	"github.com/function61/hautomo/pkg/evdevcodes"
	"github.com/function61/hautomo/pkg/ezstack"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
	"github.com/function61/hautomo/pkg/ezstack/zigbee"
)

// unique id is ZigbeeDevice.IEEEAddress
type Device struct {
	FriendlyName string          `json:"friendly_name"`
	Area         string          `json:"area"`          // added to Home Assistant only if this is set
	ZigbeeDevice *ezstack.Device `json:"zigbee_device"` // zigbee-level details
	State        *DeviceState    `json:"state"`
}

type DeviceState struct {
	LinkQuality   *AttrInt                          `json:"link_quality,omitempty"` // signal quality, "LQI" 0-255
	EndpointAttrs map[zigbee.EndpointId]*Attributes `json:"attrs"`                  // endpoint-specific attrs
}

// used to also compare which new attributes we received compared to last state we knew
func (d *DeviceState) LastHeard() time.Time {
	return d.LinkQuality.LastReport // Link quality is updated each time we hear from the device
}

// you can use this if device doesn't implement multiple endpoints
func (d *Device) DefaultEndpoint() *Attributes {
	// usually the "main" (only) Zigbee Endpoint ID is 1
	return d.State.EndpointAttrs[ezstack.DefaultSingleEndpointId] // TODO: resolve dynamically?
}

func (d *Device) ImplementsCluster(cluster cluster.ClusterId) bool {
	// FIXME: bad to assume each endpoint has similar cluster support
	if len(d.ZigbeeDevice.Endpoints) > 0 {
		for _, candidate := range d.ZigbeeDevice.Endpoints[0].InClusterList {
			if candidate == cluster {
				return true
			}
		}
	}

	return false
}

// encapsulates device endpoint's (a device can have many) attribute values like sensor values (temperature, humidity etc.)
// TODO: rename to endpointattributes?
type Attributes struct {
	Presence         *AttrBool    `json:"presence,omitempty"`   // user is present in a room
	Brightness       *AttrInt     `json:"brightness,omitempty"` // 0-255
	Color            *AttrColorXY `json:"color,omitempty"`      // color in (X,Y)
	ColorTemperature *AttrInt     `json:"color_temp,omitempty"` // mireds
	On               *AttrBool    `json:"on,omitempty"`         // on/off state
	Press            *AttrPress   `json:"press,omitempty"`      // generalized version of push
	Contact          *AttrBool    `json:"contact,omitempty"`    // true=> door/window closed
	WaterDetected    *AttrBool    `json:"water_detected,omitempty"`
	Vibration        *AttrEvent   `json:"vibration,omitempty"`
	Tilt             *AttrEvent   `json:"tilt,omitempty"`
	AlertSelect      *AttrEvent   `json:"alert_select,omitempty"` // light blinks
	Drop             *AttrEvent   `json:"drop,omitempty"`
	BatteryVoltage   *AttrFloat   `json:"battery_voltage,omitempty"` // [V]
	// BatteryLevel   *AttrFloat       `json:"battery_level,omitempty"` // [0-100 %]. only set if reported by device. calculated % levels to Home Assistant are usually calculated on-the-fly
	Temperature      *AttrFloat       `json:"temperature,omitempty"`  // [°C]
	HumidityRelative *AttrFloat       `json:"humidity_rel,omitempty"` // [0-100 %]
	Pressure         *AttrFloat       `json:"pressure,omitempty"`     // [hPa]
	Illuminance      *AttrFloat       `json:"illuminance,omitempty"`  // lux?
	Orientation      *AttrOrientation `json:"orientation,omitempty"`
	ShadePosition    *AttrInt         `json:"shade_position,omitempty"` // 0-100. 100 % = covers whole window, i.e. closed
	ShadeStop        *AttrEvent       `json:"shade_stop,omitempty"`

	PlaybackControl *AttrPlaybackControl `json:"playback_control,omitempty"`

	// the below maps are luckily new'd when JSON Unmarshal()'d

	CustomFloat  map[string]*AttrFloat  `json:"custom_float,omitempty"`
	CustomString map[string]*AttrString `json:"custom_string,omitempty"`
}

func NewAttributes() *Attributes {
	return &Attributes{
		CustomFloat:  map[string]*AttrFloat{},
		CustomString: map[string]*AttrString{},
	}
}

func (a *Attributes) CopyDifferentAttrsTo(dest *Attributes) {
	// NOTE: CopyIfDifferent() is safe to call even if both source and destination attrs are nil
	source := a

	source.On.CopyIfDifferent(&dest.On)
	source.Brightness.CopyIfDifferent(&dest.Brightness)
	source.Color.CopyIfDifferent(&dest.Color)
	source.ColorTemperature.CopyIfDifferent(&dest.ColorTemperature)
	source.ShadePosition.CopyIfDifferent(&dest.ShadePosition)

	// TODO: put into event struct
	eventCopyIfDifferent := func(source *AttrEvent, dest **AttrEvent) {
		if source != nil { // see if we even got anything to copy
			if *dest == nil || !source.LastReport.Equal((*dest).LastReport) {
				*dest = source
			}
		}
	}

	eventCopyIfDifferent(source.ShadeStop, &dest.ShadeStop)
	eventCopyIfDifferent(source.AlertSelect, &dest.AlertSelect)

	// TODO: rest. the above mainly used for writable attrs (that can be controlled from Home Assistant etc).
}

// things each attribute is expected to implement
type Attribute interface {
	LastChange() time.Time
}

type AttrString struct {
	Value      string    `json:"value"`
	LastReport time.Time `json:"reported"`
}

var _ Attribute = (*AttrString)(nil)

func (a *AttrString) LastChange() time.Time { return a.LastReport }

func (a *AttrString) CopyIfDifferent(dest **AttrString) {
	if a != nil && (*dest == nil || (*dest).Value != a.Value) {
		*dest = a
	}
}

type AttrFloat struct {
	Value      float64   `json:"value"`
	LastReport time.Time `json:"reported"`
}

var _ Attribute = (*AttrFloat)(nil)

func (a *AttrFloat) LastChange() time.Time { return a.LastReport }

func (a *AttrFloat) CopyIfDifferent(dest **AttrFloat) {
	if a != nil && (*dest == nil || (*dest).Value != a.Value) {
		*dest = a
	}
}

type AttrInt struct {
	Value      int64     `json:"value"`
	LastReport time.Time `json:"reported"`
}

var _ Attribute = (*AttrInt)(nil)

func (a *AttrInt) LastChange() time.Time { return a.LastReport }

func (a *AttrInt) CopyIfDifferent(dest **AttrInt) {
	if a != nil && (*dest == nil || (*dest).Value != a.Value) {
		*dest = a
	}
}

type AttrBool struct {
	Value      bool      `json:"value"`
	LastReport time.Time `json:"reported"`
}

var _ Attribute = (*AttrBool)(nil)

func (a *AttrBool) LastChange() time.Time { return a.LastReport }

func (a *AttrBool) CopyIfDifferent(dest **AttrBool) {
	if a != nil && (*dest == nil || (*dest).Value != a.Value) {
		*dest = a
	}
}

// event is like bool, but with "always true", i.e. it just happened.
// there is no "not happened" - or if there is it implicitly at times when even didn't happen.
type AttrEvent struct {
	LastReport time.Time `json:"reported"`
}

var _ Attribute = (*AttrEvent)(nil)

func (a *AttrEvent) LastChange() time.Time { return a.LastReport }

type AttrOrientation struct {
	X          int
	Y          int
	Z          int
	LastReport time.Time `json:"reported"`
}

var _ Attribute = (*AttrOrientation)(nil)

func (a *AttrOrientation) LastChange() time.Time { return a.LastReport }

type AttrColorXY struct {
	X          uint16
	Y          uint16
	LastReport time.Time `json:"reported"`
}

var _ Attribute = (*AttrColorXY)(nil)

func (a *AttrColorXY) LastChange() time.Time { return a.LastReport }

func (a *AttrColorXY) CopyIfDifferent(dest **AttrColorXY) {
	if a != nil && (*dest == nil || (*dest).X != a.X || (*dest).Y != a.Y) {
		*dest = a
	}
}

type PressKind uint8

const (
	PressKindUp   PressKind = 0 // untraditionally keyup first so we can "omitempty" the zero value
	PressKindDown PressKind = 1
	PressKindHold PressKind = 2
)

// Models a key/button press. Supports key combinations and double/triple/etc. clicks. Examples:
//
// {Key: KeyPOWER} for a power toggle
// {Key: Btn0, KeysAdditional=[Btn1], Kind=Hold} for two buttons held in a switch that has two buttons
// {Key: Btn0, Times=2} for a double click in a single-button switch
type AttrPress struct {
	// key press, identified by Linux evdev code. examples (below Key{VOLUMEUP} means evdevcodes.KeyVOLUMEUP):
	// - Key{BRIGHTNESSUP, BRIGHTNESSDOWN} if key is for ± brightness
	// - Key{VOLUMEUP, VOLUMEDOWN} if key is for ± volume
	// - Key{OK, SELECT, CANCEL, BACK} for generic selections and navigation
	// - Key{POWER} if key is for power (toggling) generic devices (not just lights)
	// - Key{OPEN, CLOSE} if key is for power opening or closing something (curtains, shades, door etc.)
	// - Key{LIGHTS_TOGGLE} if key can be used only for toggling lights (like if the key has a light symbol instead of a power symbol)
	// - Btn{0,...,9} for generic buttons or un-numbered buttons whose press isn't intended to produce their number: "button 1", "button 2", ... or "leftmost button", "the second button", "third button" etc.
	// - Key{NUMERIC_0,...NUMERIC_9,NUMERIC_STAR,NUMERIC_POUND} for keypads, remote controls **where the buttons have number labels**
	// - Key{NEXT, PREVIOUS} if key is for selecting a list item, be it playlist, scene list or similar
	// - Key{NEXTSONG, PREVIOUSSONG} if key is for controlling audio. prefer {NEXT,PREVIOUS} if playlist can contain video or other non-audio items.
	// - Key{PLAYPAUSE, PLAY, STOP, MUTE, etc.} if key is for media control
	// - Key{LEFT, RIGHT} if key has horizontal directionality and no distinct evdev property to control (volume, brightness)
	// - Key{UP, DOWN} if key has vertical directionality and no distinct evdev property to control (volume, brightness)
	// - Most suitable evdev code if above instructions don't cover your use case
	Key            evdevcodes.KeyOrButton   `json:"key"`                       // primary (or first) key. for convenience because usually we get only single-key presses
	KeysAdditional []evdevcodes.KeyOrButton `json:"keys_additional,omitempty"` // in rare occasion there are multiple keys pressed at once. additional keys to primary are here. see AllKeys()
	Kind           PressKind                `json:"kind,omitempty"`            // usually PressKindUp, but can be down/hold if we have hold detection or get separate keyup, keydown events
	CountRaw       *int                     `json:"count,omitempty"`           // double, triple etc. clicks. only specify if >= 2. see Count()

	LastReport time.Time `json:"reported"`
}

// it is recommended to use this to know all keys which were pressed instead of accessing Key/KeysAdditional directly!
func (a *AttrPress) AllKeys() []evdevcodes.KeyOrButton {
	return append([]evdevcodes.KeyOrButton{a.Key}, a.KeysAdditional...)
}

func (a *AttrPress) HasKey(key evdevcodes.KeyOrButton) bool {
	for _, item := range a.AllKeys() {
		if item == key {
			return true
		}
	}

	return false
}

// returns how many times they key was pressed
func (a *AttrPress) Count() int {
	if a.CountRaw != nil {
		return *a.CountRaw
	} else {
		return 1
	}
}

var _ Attribute = (*AttrPress)(nil)

func (a *AttrPress) LastChange() time.Time { return a.LastReport }

type PlaybackControl string

const (
	PlaybackControlPlay        PlaybackControl = "Play"
	PlaybackControlPause       PlaybackControl = "Pause"
	PlaybackControlStop        PlaybackControl = "Stop"
	PlaybackControlStartOver   PlaybackControl = "StartOver"
	PlaybackControlPrevious    PlaybackControl = "Previous"
	PlaybackControlNext        PlaybackControl = "Next"
	PlaybackControlRewind      PlaybackControl = "Rewind"
	PlaybackControlFastForward PlaybackControl = "FastForward"
)

type AttrPlaybackControl struct {
	Control    PlaybackControl `json:"control"`
	LastReport time.Time       `json:"reported"`
}

func (a *AttrPlaybackControl) LastChange() time.Time { return a.LastReport }

var _ Attribute = (*AttrPlaybackControl)(nil)

// builder helper for wrapping attributes with shared timestamp
type AttrsCtx struct {
	Attrs    *Attributes
	Reported time.Time
	Endpoint zigbee.EndpointId
}

func (a *AttrsCtx) Int(value int64) *AttrInt {
	return &AttrInt{
		Value:      value,
		LastReport: a.Reported,
	}
}

func (a *AttrsCtx) Float(value float64) *AttrFloat {
	return &AttrFloat{
		Value:      value,
		LastReport: a.Reported,
	}
}

func (a *AttrsCtx) Bool(value bool) *AttrBool {
	return &AttrBool{
		Value:      value,
		LastReport: a.Reported,
	}
}

func (a *AttrsCtx) String(value string) *AttrString {
	return &AttrString{
		Value:      value,
		LastReport: a.Reported,
	}
}

func (a *AttrsCtx) Event() *AttrEvent {
	return &AttrEvent{
		LastReport: a.Reported,
	}
}

func (a *AttrsCtx) ColorXY(x, y uint16) *AttrColorXY {
	return &AttrColorXY{
		X:          x,
		Y:          y,
		LastReport: a.Reported,
	}
}

func (a *AttrsCtx) PressUp(key evdevcodes.KeyOrButton) *AttrPress {
	return &AttrPress{
		Key:        key,
		Kind:       PressKindUp,
		LastReport: a.Reported,
	}
}

func (a *AttrsCtx) PlaybackControl(control PlaybackControl) *AttrPlaybackControl {
	return &AttrPlaybackControl{
		Control:    control,
		LastReport: a.Reported,
	}
}
