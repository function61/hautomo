package happylights

import (
	"github.com/function61/gokit/assert"
	"testing"
)

func TestRequestToHex(t *testing.T) {
	addr := "A4:C1:38:CC:AC:85"

	assert.EqualString(t, requestToHex(RequestOn(addr)), "cc2333")
	assert.EqualString(t, requestToHex(RequestOff(addr)), "cc2433")

	assert.EqualString(t, requestToHex(RequestRGB(addr, 255, 255, 255)), "56ffffff00f0aa")
	assert.EqualString(t, requestToHex(RequestRGB(addr, 255, 0, 0)), "56ff000000f0aa")
	assert.EqualString(t, requestToHex(RequestRGB(addr, 0, 255, 0)), "5600ff0000f0aa")
	assert.EqualString(t, requestToHex(RequestRGB(addr, 0, 0, 255)), "560000ff00f0aa")
	assert.EqualString(t, requestToHex(RequestRGB(addr, 0, 0, 128)), "5600008000f0aa")

	assert.EqualString(t, requestToHex(RequestWhite(addr, 0)), "56000000000faa")
	assert.EqualString(t, requestToHex(RequestWhite(addr, 96)), "56000000600faa")
	assert.EqualString(t, requestToHex(RequestWhite(addr, 255)), "56000000ff0faa")
}
