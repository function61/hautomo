package main

import (
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/pkg/adapters/alexaadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/devicegroupadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/dummyadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/eventghostnetworkclientadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/harmonyhubadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/ikeatradfriadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/irsimulatoradapter"
	"github.com/function61/home-automation-hub/pkg/adapters/lircadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/particleadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/presencebypingadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/trionesadapter"
	"github.com/function61/home-automation-hub/pkg/adapters/zigbee2mqttadapter"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
)

type AdapterInitFn func(adapter *hapitypes.Adapter, stop *stopper.Stopper) error

var adapters = map[string]AdapterInitFn{
	"devicegroup":             devicegroupadapter.Start,
	"dummy":                   dummyadapter.Start,
	"eventghostnetworkclient": eventghostnetworkclientadapter.Start,
	"triones":                 trionesadapter.Start,
	"harmony":                 harmonyhubadapter.Start,
	"ikea_tradfri":            ikeatradfriadapter.Start,
	"zigbee2mqtt":             zigbee2mqttadapter.Start,
	"irsimulator":             irsimulatoradapter.Start,
	"lirc":                    lircadapter.Start,
	"particle":                particleadapter.Start,
	"presencebyping":          presencebypingadapter.Start,
	"sqs":                     alexaadapter.Start,
}
