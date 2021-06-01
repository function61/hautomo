package deviceadapters

import (
	"testing"

	"github.com/function61/gokit/testing/assert"
)

func TestParseAttributeList(t *testing.T) {
	inputPayload := []byte{0x01, 0x21, 0x9f, 0x0b, 0x04, 0x21, 0xa8, 0x13, 0x05, 0x21,
		0x2d, 0x00, 0x06, 0x24, 0x02, 0x00, 0x00, 0x00, 0x00, 0x64,
		0x29, 0x2c, 0x07, 0x65, 0x21, 0x64, 0x11, 0x0a, 0x21, 0xe1, 0x76}

	attrs, err := ParseAttributeList(inputPayload)
	assert.Ok(t, err)

	assert.EqualJson(t, len(attrs), "7")
	assert.Assert(t, attrs[0].Id == 1)
	assert.Assert(t, attrs[0].Attribute.Value.(uint64) == 2975)

	assert.Assert(t, attrs[4].Id == 100)
	assert.Assert(t, attrs[4].Attribute.Value.(int64) == 1836)

	assert.Assert(t, attrs.Find(100).Value.(int64) == 1836)
}

// search for RTCGQ11LM_interval and 65281
// https://github.com/Koenkk/zigbee-herdsman-converters/blob/2792fb0e1722ca874d3f43d1f688778661c189a1/converters/fromZigbee.js
