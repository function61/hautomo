package suntimes

import (
	"testing"
	"time"

	"github.com/function61/gokit/assert"
)

func TestIsBetweenGoldenHours(t *testing.T) {
	helsinki, err := time.LoadLocation("Europe/Helsinki")
	assert.Assert(t, err == nil)

	const testDateFormat = "2006-01-02 15:04"

	tests := []struct {
		time          string
		expectedLight bool
	}{
		// january in Finland is shit
		{"2019-01-14 00:00", false},
		{"2019-01-14 03:00", false},
		{"2019-01-14 09:00", false},
		{"2019-01-14 11:26", false},
		{"2019-01-14 11:27", true},
		{"2019-01-14 12:00", true},
		{"2019-01-14 13:43", true},
		{"2019-01-14 13:44", false},
		{"2019-01-14 19:00", false},
		{"2019-01-14 23:59", false},

		// summer in Finland is awesome
		{"2019-07-01 01:00", false},
		{"2019-07-01 03:00", false},
		{"2019-07-01 05:16", false},
		{"2019-07-01 05:17", true},
		{"2019-07-01 06:00", true},
		{"2019-07-01 10:00", true},
		{"2019-07-01 18:00", true},
		{"2019-07-01 21:30", true},
		{"2019-07-01 22:00", false},
		{"2019-07-01 23:00", false},
		{"2019-07-01 23:59", false},
	}

	for _, test := range tests {
		test := test // pin

		t.Run(test.time, func(t *testing.T) {
			now, err := time.ParseInLocation(testDateFormat, test.time, helsinki)
			assert.Assert(t, err == nil)

			light := IsBetweenGoldenHours(now, Tampere)

			assert.Assert(t, light == test.expectedLight)
		})
	}
}
