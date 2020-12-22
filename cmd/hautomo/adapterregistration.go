package main

import (
	"context"

	"github.com/function61/hautomo/pkg/adapters/alexaadapter"
	"github.com/function61/hautomo/pkg/adapters/devicegroupadapter"
	"github.com/function61/hautomo/pkg/adapters/dummyadapter"
	"github.com/function61/hautomo/pkg/adapters/eventghostadapter"
	"github.com/function61/hautomo/pkg/adapters/harmonyhubadapter"
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
	"triones":        trionesadapter.Start,
	"harmony":        harmonyhubadapter.Start,
	"ikea_tradfri":   ikeatradfriadapter.Start,
	"zigbee2mqtt":    zigbee2mqttadapter.Start,
	"irsimulator":    irsimulatoradapter.Start,
	"lirc":           lircadapter.Start,
	"particle":       particleadapter.Start,
	"presencebyping": presencebypingadapter.Start,
	"sonoff":         sonoffadapter.Start,
	"screen-server":  screenserveradapter.Start,
	"sqs":            alexaadapter.Start,
}
