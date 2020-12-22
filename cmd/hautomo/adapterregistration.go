package main

import (
	"context"

	"github.com/function61/hautomo/pkg/adapters/alexaadapter"
	"github.com/function61/hautomo/pkg/adapters/devicegroupadapter"
	"github.com/function61/hautomo/pkg/adapters/dummyadapter"
	"github.com/function61/hautomo/pkg/adapters/eventghostadapter"
	"github.com/function61/hautomo/pkg/adapters/harmonyhubadapter"
	"github.com/function61/hautomo/pkg/adapters/homeassistantadapter"
	"github.com/function61/hautomo/pkg/adapters/ikeatradfriadapter"
	"github.com/function61/hautomo/pkg/adapters/irsimulatoradapter"
	"github.com/function61/hautomo/pkg/adapters/lircadapter"
	"github.com/function61/hautomo/pkg/adapters/particleadapter"
	"github.com/function61/hautomo/pkg/adapters/presencebypingadapter"
	"github.com/function61/hautomo/pkg/adapters/screenserveradapter"
	"github.com/function61/hautomo/pkg/adapters/sonoffadapter"
	"github.com/function61/hautomo/pkg/adapters/trionesadapter"
	"github.com/function61/hautomo/pkg/adapters/zigbee2mqttadapter"
	"github.com/function61/hautomo/pkg/hapitypes"
)

type AdapterInitFn func(ctx context.Context, adapter *hapitypes.Adapter) error

var adapters = map[string]AdapterInitFn{
	"devicegroup":    devicegroupadapter.Start,
	"dummy":          dummyadapter.Start,
	"eventghost":     eventghostadapter.Start,
	"harmony":        harmonyhubadapter.Start,
	"home-assistant": homeassistantadapter.Start,
	"ikea_tradfri":   ikeatradfriadapter.Start,
	"irsimulator":    irsimulatoradapter.Start,
	"lirc":           lircadapter.Start,
	"particle":       particleadapter.Start,
	"presencebyping": presencebypingadapter.Start,
	"screen-server":  screenserveradapter.Start,
	"sonoff":         sonoffadapter.Start,
	"sqs":            alexaadapter.Start,
	"triones":        trionesadapter.Start,
	"zigbee2mqtt":    zigbee2mqttadapter.Start,
}
