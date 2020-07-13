package hapitypes

import (
	"testing"

	"github.com/function61/gokit/assert"
)

func TestRGBIsGrayscale(t *testing.T) {
	assert.Assert(t, NewRGB(255, 255, 255).IsGrayscale() == true)
	assert.Assert(t, NewRGB(0, 0, 0).IsGrayscale() == true)

	assert.Assert(t, NewRGB(255, 0, 0).IsGrayscale() == false)
	assert.Assert(t, NewRGB(0, 255, 0).IsGrayscale() == false)
	assert.Assert(t, NewRGB(0, 0, 255).IsGrayscale() == false)
	assert.Assert(t, NewRGB(255, 255, 254).IsGrayscale() == false)
}
