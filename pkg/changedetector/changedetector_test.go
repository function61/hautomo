package changedetector

import (
	"strings"
	"testing"

	"github.com/function61/gokit/testing/assert"
)

func TestChangeDetector(t *testing.T) {
	detector := New()

	changed, err := detector.ReaderChanged(strings.NewReader("why hello there"))
	assert.Ok(t, err)

	assert.Assert(t, changed)

	changed, err = detector.ReaderChanged(strings.NewReader("why hello there"))
	assert.Ok(t, err)

	assert.Assert(t, !changed)

	changed, err = detector.ReaderChanged(strings.NewReader("#why hello there"))
	assert.Ok(t, err)

	assert.Assert(t, changed)
}
