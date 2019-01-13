package lircadapter

import (
	"github.com/function61/gokit/assert"
	"testing"
)

func TestIrwOutputLineToIrEvent(t *testing.T) {
	tests := []struct {
		input              string
		expectedRemoteName string
		expectedKey        string
	}{
		{
			input:              "000000037ff07bee 00 KEY_VOLUMEDOWN devinput",
			expectedRemoteName: "devinput",
			expectedKey:        "KEY_VOLUMEDOWN",
		},
		{
			input:              "0000000000f40bf0 00 KEY_POWER mceusb",
			expectedRemoteName: "mceusb",
			expectedKey:        "KEY_POWER",
		},
		{
			input:              "0000000f40bf0 00 KEY_POWER mceusb", // too short prefix - should not parse
			expectedRemoteName: "",
			expectedKey:        "",
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			evt := irwOutputLineToIrEvent(test.input)

			if test.expectedRemoteName != "" {
				assert.EqualString(t, evt.Remote, test.expectedRemoteName)
				assert.EqualString(t, evt.Event, test.expectedKey)
			} else {
				assert.True(t, evt == nil)
			}
		})
	}
}
