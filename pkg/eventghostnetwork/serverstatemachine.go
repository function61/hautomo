package eventghostnetwork

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
)

const (
	magicKnock = "quintessence"
)

type eventListener func(event string, payload []string, password string)

type serverState int

const (
	serverStateWaitingMagicKnock serverState = iota
	serverStateWaitingChallengeResponse
	serverStateAuthenticated
	serverStateDisconnected
)

var stateHandlers = map[serverState]func(*serverStateMachine, string) (string, serverState){
	serverStateWaitingMagicKnock:        (*serverStateMachine).handleMagicKnock,
	serverStateWaitingChallengeResponse: (*serverStateMachine).handleChallengeResponse,
	serverStateAuthenticated:            (*serverStateMachine).handleProcessEvent,
	serverStateDisconnected:             (*serverStateMachine).handleDisconnected,
}

type serverStateMachine struct {
	state                      serverState
	passwordChallengeResponses map[string]string // challenge => correct password
	authenticatedPassword      string
	cookie                     string
	payload                    []string
	eventListener              eventListener
}

func newServerStateMachine(passwords []string, listener eventListener) *serverStateMachine {
	cookieBytes := make([]byte, 16)
	if _, err := rand.Read(cookieBytes); err != nil {
		panic(err)
	}
	cookie := base64.RawStdEncoding.EncodeToString(cookieBytes)

	passwordChallengeResponses := make(map[string]string, len(passwords))
	for _, password := range passwords {
		passwordChallengeResponses[calculateExpectedChallengeResponse(cookie, password)] = password
	}

	return &serverStateMachine{
		state:                      serverStateWaitingMagicKnock,
		passwordChallengeResponses: passwordChallengeResponses,
		cookie:                     cookie,
		payload:                    []string{},
		eventListener:              listener,
	}
}

func (s *serverStateMachine) process(input string) string {
	handler, exists := stateHandlers[s.state]
	if !exists {
		panic("unknown state") // should not hapen
	}

	response, nextState := handler(s, input)
	s.state = nextState

	return response
}

func (s *serverStateMachine) handleMagicKnock(input string) (string, serverState) {
	if input != magicKnock {
		return "", serverStateDisconnected
	}

	return s.cookie, serverStateWaitingChallengeResponse
}

func (s *serverStateMachine) handleChallengeResponse(challengeResponse string) (string, serverState) {
	password, found := s.passwordChallengeResponses[challengeResponse]
	if !found {
		return "", serverStateDisconnected
	}

	s.authenticatedPassword = password

	return "accept", serverStateAuthenticated
}

func (s *serverStateMachine) handleProcessEvent(input string) (string, serverState) {
	if input == "close" {
		return "", serverStateDisconnected
	}

	if strings.HasPrefix(input, "payload ") {
		s.payload = append(s.payload, input[len("payload "):])

		return "", serverStateAuthenticated
	}

	s.eventListener(input, s.payload, s.authenticatedPassword)

	s.payload = []string{}

	return "", serverStateAuthenticated
}

func (s *serverStateMachine) handleDisconnected(input string) (string, serverState) {
	// nothing should happen after we are disconnected (this method shouldn't even be called)
	return "", serverStateDisconnected
}

func calculateExpectedChallengeResponse(cookie string, password string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(cookie+":"+password)))
}
