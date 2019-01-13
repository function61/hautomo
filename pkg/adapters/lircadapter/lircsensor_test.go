package lircadapter

import (
	"github.com/function61/gokit/assert"
	"testing"
)

func TestIrwOutputLineToIrEvent(t *testing.T) {
	evt := irwOutputLineToIrEvent("000000037ff07bee 00 KEY_VOLUMEDOWN devinput")
	assert.EqualString(t, evt.Remote, "mceusb")
	assert.EqualString(t, evt.Event, "KEY_VOLUMEDOWN")
}
