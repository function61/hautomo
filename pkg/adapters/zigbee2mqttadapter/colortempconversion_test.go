package zigbee2mqttadapter

import (
	"github.com/function61/gokit/assert"
	"testing"
)

func TestKelvinToMired(t *testing.T) {
	assert.Assert(t, kelvinToMired(2200) == 454) // warm, warm white
	assert.Assert(t, kelvinToMired(2700) == 370) // incandescent, soft white
	assert.Assert(t, kelvinToMired(4000) == 250) // white
	assert.Assert(t, kelvinToMired(5500) == 181) // daylight, daylight white
	assert.Assert(t, kelvinToMired(7000) == 142) // cool, cool white
}
