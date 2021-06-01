package deviceadapters

import (
	"errors"

	"github.com/function61/hautomo/pkg/ezstack"
	"github.com/function61/hautomo/pkg/ezstack/ezhub/hubtypes"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
)

var (
	errUnhandledCommand = errors.New("unhandled command")
)

// helper
func For(dev *ezstack.Device) Adapter {
	return AdapterForModel(dev.Model)
}

func AdapterForModel(model ezstack.Model) Adapter {
	if adapter, found := adapterByModel[model]; found {
		return adapter
	}

	return adapterByModel["*"] // catch-all
}

func defineAdapter(model ezstack.Model, opts ...optFn) {
	adapterByModel[model] = &mergedDevice{newAdapter(opts...), adapterByModel["*"]}
}

type Adapter interface {
	// given "genOnOff.onOff", returns parser for it
	ParserForAttribute(clusterAndAttribute string) customParser

	// handle any additional (non-AttributeReport) commands
	HandleCommand(command interface{}, actx *hubtypes.AttrsCtx) error

	// returns nil if not battery powered
	BatteryType() *hubtypes.BatteryType
}

// swallows an attribute we know we don't need (so it is not logged as surprising unsupported attribute)
// e.g. is some heartbeat messages sending their device model ID
func noopParser(_ *cluster.Attribute, _ *hubtypes.AttrsCtx) error {
	return nil
}

// parser is responsible for assigning data from Zigbee message to our internal data model.
// e.g. Zigbee sends genOnOff.onOff=bool(true) -> we assign it to Attrs.On.Value, but if it's
// an Aqara button the same message means Attrs.Pushes.Value = 1 (for a single-click)
type customParser func(*cluster.Attribute, *hubtypes.AttrsCtx) error

type attributeMatcher func(string) customParser

type optFn func(*simpleAdapter)

func attributeParser(clusterAndAttribute string, parser customParser) optFn {
	return func(adapter *simpleAdapter) {
		adapter.matchers = append(adapter.matchers, func(clusterAndAttributeCandidate string) customParser {
			if clusterAndAttributeCandidate == clusterAndAttribute {
				return parser
			} else {
				return nil
			}
		})
	}
}

func withBatteryType(batteryType *hubtypes.BatteryType) optFn {
	return func(adapter *simpleAdapter) {
		adapter.batteryType = batteryType
	}
}

func withCommandHandler(handler commandHandlerFn) optFn {
	return func(adapter *simpleAdapter) {
		adapter.commandHandler = handler
	}
}

type commandHandlerFn func(command interface{}, actx *hubtypes.AttrsCtx) error

// covers the most common use case where we pass attribute to a function for parsing
type simpleAdapter struct {
	matchers       []attributeMatcher
	batteryType    *hubtypes.BatteryType
	commandHandler commandHandlerFn
}

func newAdapter(opts ...optFn) Adapter {
	adapter := &simpleAdapter{}
	for _, opt := range opts {
		opt(adapter)
	}

	return adapter
}

var _ Adapter = (*simpleAdapter)(nil)

func (h *simpleAdapter) ParserForAttribute(clusterAndAttribute string) customParser {
	for _, matcher := range h.matchers {
		if parser := matcher(clusterAndAttribute); parser != nil {
			return parser
		}
	}

	return nil
}

func (h *simpleAdapter) HandleCommand(command interface{}, actx *hubtypes.AttrsCtx) error {
	if h.commandHandler != nil {
		return h.commandHandler(command, actx)
	} else {
		return errUnhandledCommand
	}
}

func (h *simpleAdapter) BatteryType() *hubtypes.BatteryType {
	return h.batteryType
}

// uses primary device (model-specific parsers) to find a parser for an attribute.
// if no such thing is found, returns parser from secondary device.
//
// used as building block to transparently paper over differences, e.g. meanings:
//
//           usually: genOnOff.onOff=true -> turn on
// with Aqara Button: genOnOff.onOff=true -> single click
type mergedDevice struct {
	primary   Adapter
	secondary Adapter
}

var _ Adapter = (*mergedDevice)(nil)

func (m *mergedDevice) ParserForAttribute(clusterAndAttribute string) customParser {
	if parser := m.primary.ParserForAttribute(clusterAndAttribute); parser != nil {
		return parser
	} else { // no "overridden" parser found => use default parser
		return m.secondary.ParserForAttribute(clusterAndAttribute)
	}
}

func (m *mergedDevice) HandleCommand(command interface{}, actx *hubtypes.AttrsCtx) error {
	// first give the primary a chance to handle it
	err := m.primary.HandleCommand(command, actx)
	if err == errUnhandledCommand { // didn't handle => give the secondary a chance
		err = m.secondary.HandleCommand(command, actx)
	}

	return err
}

func (m *mergedDevice) BatteryType() *hubtypes.BatteryType {
	return m.primary.BatteryType()
}
