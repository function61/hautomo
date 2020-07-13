package eventghostnetwork

import (
	"strings"
	"testing"

	"github.com/function61/gokit/assert"
)

var noEvents = func(event string, payload []string, password string) { panic("should not receive any event") }

var testingPassword = []string{"testingPassword"}

func TestPostTwoEvents(t *testing.T) {
	lastEvent := ""

	srv := newServerStateMachine(testingPassword, func(event string, payload []string, password string) {
		assert.EqualString(t, password, testingPassword[0])

		lastEvent = event + " " + strings.Join(payload, "|")
	})

	assert.Assert(t, srv.state == serverStateWaitingMagicKnock)
	cookie := srv.process("quintessence")

	assert.Assert(t, srv.state == serverStateWaitingChallengeResponse)
	assert.EqualString(t, cookie, srv.cookie)

	assert.EqualString(t, srv.process(calculateExpectedChallengeResponse(cookie, "testingPassword")), "accept")
	assert.Assert(t, srv.state == serverStateAuthenticated)

	assert.EqualString(t, srv.process("event1"), "")
	assert.EqualString(t, lastEvent, "event1 ")
	assert.Assert(t, srv.state == serverStateAuthenticated)

	assert.EqualString(t, srv.process("payload foo"), "")
	assert.EqualString(t, srv.process("payload bar"), "")
	assert.EqualString(t, srv.process("event2"), "")
	assert.EqualString(t, lastEvent, "event2 foo|bar")
	assert.Assert(t, srv.state == serverStateAuthenticated)

	assert.EqualString(t, srv.process("close"), "")
	assert.Assert(t, srv.state == serverStateDisconnected)
}

func TestInvalidKnock(t *testing.T) {
	srv := newServerStateMachine(testingPassword, noEvents)

	assert.EqualString(t, srv.process("wrongKnock"), "")
	assert.Assert(t, srv.state == serverStateDisconnected)
}

func TestInvalidAuth(t *testing.T) {
	srv := newServerStateMachine(testingPassword, noEvents)

	cookie := srv.process(magicKnock)
	assert.Assert(t, srv.state == serverStateWaitingChallengeResponse)

	assert.EqualString(t, srv.process(calculateExpectedChallengeResponse(cookie, "badPassword")), "")
	assert.Assert(t, srv.state == serverStateDisconnected)
}
