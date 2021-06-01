package homeassistantmqtt

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/function61/hautomo/pkg/evdevcodes"
	"github.com/function61/hautomo/pkg/ezstack/ezhub/hubtypes"
)

const (
	xyColorScaleMax = 65279
)

// generic = able to represent pretty much any zigbee2mqtt device, in both directions (setting things and notifying of changes)
// https://www.zigbee2mqtt.io/information/mqtt_topics_and_message_structure.html
type zigbee2mqttGenericJson struct {
	State       *string  `json:"state,omitempty"`
	Occupancy   *bool    `json:"occupancy,omitempty"`
	Contact     *bool    `json:"contact,omitempty"`
	Illuminance *int     `json:"illuminance,omitempty"`
	Action      *string  `json:"action,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
	Humidity    *float64 `json:"humidity,omitempty"` // [0-100 %]
	Pressure    *float64 `json:"pressure,omitempty"`
	Brightness  *uint8   `json:"brightness,omitempty"`
	Color       *color   `json:"color,omitempty"`
	ColorTemp   *uint16  `json:"color_temp,omitempty"`
	WaterLeak   *bool    `json:"water_leak,omitempty"`

	Transition *int `json:"transition,omitempty"` // transition duration (in 100ms?)

	AngleX *int `json:"angle_x,omitempty"`
	AngleY *int `json:"angle_y,omitempty"`
	AngleZ *int `json:"angle_z,omitempty"`

	Alert *string `json:"alert,omitempty"` // usually "select"

	LinkQuality *int64            `json:"linkquality,omitempty"` // [LQI]
	Voltage     *int64            `json:"voltage,omitempty"`     // [mV]
	Battery     *int64            `json:"battery,omitempty"`     // [%]
	Extra       map[string]string `json:"extra,omitempty"`

	HackShadeCommand *string `json:"shade_command,omitempty"` // not really in Home Assistant
}

func MessageFromChangedAttributes(
	attrs *hubtypes.Attributes,
	dev *hubtypes.Device,
	batteryType *hubtypes.BatteryType,
	now time.Time,
) (string, error) {
	// Home Asssistant wants to see the full state of sensor values each time.
	// action values not. so for sensor values we use known(), and reportedNow() for action values

	// some attributes we need to always send, even if they didn't change (Home Assistant expects this)
	known := func(attr hubtypes.Attribute) bool { // helper
		return !isNilInterface(attr)
	}

	// distinct from "changed". e.g. events can be reported over and over again, so can
	// occupancy=true, occupancy=true
	reportedNow := func(attr hubtypes.Attribute) bool { // helper
		if !known(attr) {
			return false
		}

		return attr.LastChange().Equal(now)
	}

	voltage, battery := func() (*int64, *int64) {
		if known(attrs.BatteryVoltage) {
			var batteryLevelPtr *int64
			if batteryType != nil {
				batteryLevel := int64(batteryType.VoltageToPercentage(attrs.BatteryVoltage.Value) * 100)
				batteryLevelPtr = &batteryLevel
			}

			voltage := int64(attrs.BatteryVoltage.Value * 1000) // [mV]
			return &voltage, batteryLevelPtr
		} else {
			return nil, nil
		}
	}()

	angleX, angleY, angleZ := func() (*int, *int, *int) {
		if known(attrs.Orientation) {
			ori := attrs.Orientation
			return &ori.X, &ori.Y, &ori.Z
		} else {
			return nil, nil, nil
		}
	}()

	// handle clicks
	action := func() *string {
		if reportedNow(attrs.Press) {
			// 1 => "single"
			// ...
			// 4 => "quadruple"
			// (or "hold" if holding)
			countText := func() string {
				if attrs.Press.Kind == hubtypes.PressKindHold {
					return "hold"
				} else {
					return pushesToString(attrs.Press.Count())
				}
			}()

			buttonLabel := func() string {
				left := attrs.Press.HasKey(evdevcodes.Btn0)
				right := attrs.Press.HasKey(evdevcodes.Btn1)

				switch {
				case left && right:
					return "both"
				case left:
					return "left"
				case right:
					return "right"
				default:
					return strconv.Itoa(int(attrs.Press.Key)) // NOTE: ignores additional keys
				}
			}()

			if dev.ZigbeeDevice.Model == "lumi.remote.b286acn01\x00\x00\x00" {
				// looks like "double_right"
				return stringPtr(fmt.Sprintf("%s_%s", countText, buttonLabel))
			} else {
				return stringPtr(countText)
			}
		} else {
			return nil
		}
	}()

	// action might also be vibration/drop/tilt etc. (click is expected to be nil in these cases)

	if reportedNow(attrs.Vibration) {
		action = stringPtr("vibration")
	}

	if reportedNow(attrs.Drop) {
		action = stringPtr("drop")
	}

	if reportedNow(attrs.Tilt) {
		action = stringPtr("tilt")
	}

	extra := map[string]string{}
	for key, val := range attrs.CustomString {
		extra[key] = val.Value
	}

	jsonBytes, err := json.Marshal(zigbee2mqttGenericJson{
		Action:  action,
		Voltage: voltage,
		Battery: battery,
		AngleX:  angleX,
		AngleY:  angleY,
		AngleZ:  angleZ,
		Extra:   extra,
		LinkQuality: func() *int64 {
			if dev.State.LinkQuality != nil { // can be nil mainly for testing
				return &dev.State.LinkQuality.Value
			} else {
				return nil
			}
		}(),
		State: func() *string {
			// Home Assistant requires us to always send this
			if known(attrs.On) {
				if attrs.On.Value {
					return stringPtr("ON")
				} else {
					return stringPtr("OFF")
				}
			} else {
				return nil
			}
		}(),
		Occupancy: func() *bool {
			if reportedNow(attrs.Presence) {
				return &attrs.Presence.Value
			} else {
				return nil
			}
		}(),
		Contact: func() *bool {
			if reportedNow(attrs.Contact) {
				return &attrs.Contact.Value
			} else {
				return nil
			}
		}(),
		Illuminance: func() *int {
			if known(attrs.Illuminance) {
				num := int(attrs.Illuminance.Value)
				return &num
			} else {
				return nil
			}
		}(),
		Color: func() *color {
			if reportedNow(attrs.Color) {
				return &color{
					X: float64(attrs.Color.X) / xyColorScaleMax,
					Y: float64(attrs.Color.Y) / xyColorScaleMax,
				}
			} else {
				return nil
			}
		}(),
		ColorTemp: func() *uint16 {
			if reportedNow(attrs.ColorTemperature) {
				num := uint16(attrs.ColorTemperature.Value)
				return &num
			} else {
				return nil
			}
		}(),
		Brightness: func() *uint8 {
			if reportedNow(attrs.Brightness) {
				num := uint8(attrs.Brightness.Value)
				return &num
			} else {
				return nil
			}
		}(),
		Temperature: func() *float64 {
			if known(attrs.Temperature) {
				return &attrs.Temperature.Value
			} else {
				return nil
			}
		}(),
		Humidity: func() *float64 {
			if known(attrs.HumidityRelative) {
				return &attrs.HumidityRelative.Value
			} else {
				return nil
			}
		}(),
		WaterLeak: func() *bool {
			if attrs.WaterDetected != nil {
				return &attrs.WaterDetected.Value
			} else {
				return nil
			}
		}(),
		Pressure: func() *float64 {
			if known(attrs.Pressure) {
				return &attrs.Pressure.Value
			} else {
				return nil
			}
		}(),
	})
	if err != nil { // shouldn't happen
		return "", err
	}

	return string(jsonBytes), nil
}

func MessageToAttributes(inboundMsg InboundMessage, actx *hubtypes.AttrsCtx, currentlyOn bool) error {
	attrs := actx.Attrs // shorthands
	msg := inboundMsg.Message

	if msg.State != nil {
		// some devices have toggle support, but not all I guess, and with toggle
		// we're still supposed to notify MQTT of the actual state.
		// => it's just easiest to to translate this into an explicit command so we at least
		//    know the resulting state after ACK
		if *msg.State == "TOGGLE" {
			invertedState := func() string {
				if currentlyOn { // invert
					return "OFF"
				} else {
					return "ON"
				}
			}()

			msg.State = &invertedState
		}

		switch *msg.State {
		case "ON":
			attrs.On = actx.Bool(true)
		case "OFF":
			attrs.On = actx.Bool(false)
		default:
			return fmt.Errorf("unknown on/off state: %s", *msg.State)
		}
	}

	if msg.Brightness != nil {
		attrs.Brightness = actx.Int(int64(*msg.Brightness)) // got ack => we know newest attr value
	}

	if msg.Color != nil {
		x, y, err := msg.Color.XYScaledTo65279()
		if err != nil {
			return err
		}

		attrs.Color = actx.ColorXY(x, y)
	}

	if msg.ColorTemp != nil {
		attrs.ColorTemperature = actx.Int(int64(*msg.ColorTemp))
	}

	if msg.HackShadeCommand != nil {
		switch *msg.HackShadeCommand {
		case "OPEN":
			attrs.ShadePosition = actx.Int(0)
		case "CLOSE":
			attrs.ShadePosition = actx.Int(100)
		case "STOP":
			attrs.ShadeStop = actx.Event()
		default:
			return fmt.Errorf("unknown HackShadeCommand: %s", *msg.HackShadeCommand)
		}
	}

	if msg.Alert != nil {
		switch *msg.Alert {
		case "select":
			attrs.AlertSelect = actx.Event()
		default:
			return fmt.Errorf("unknown alert: %s", *msg.Alert)
		}
	}

	return nil
}

func stringPtr(input string) *string {
	return &input
}

func isNilInterface(i interface{}) bool {
	return i == nil || reflect.ValueOf(i).IsNil()
}

var pushesToStringMap = map[int]string{
	1: "single",
	2: "double",
	3: "triple",
	4: "quadruple",
}

func pushesToString(num int) string {
	val := pushesToStringMap[num]
	if val == "" {
		panic(fmt.Errorf("pushesToString: unsupported num: %d", num))
	}

	return val
}

type color struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func (c color) XYScaledTo65279() (uint16, uint16, error) {
	if c.X > 1.0 || c.X < 0.0 {
		return 0, 0, fmt.Errorf("x point outside of range: %d", c.X)
	}

	if c.Y > 1.0 || c.Y < 0.0 {
		return 0, 0, fmt.Errorf("x point outside of range: %d", c.Y)
	}

	return uint16(c.X * xyColorScaleMax), uint16(c.Y * xyColorScaleMax), nil
}
