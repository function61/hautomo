package eventghostnetwork

import (
	"bufio"
	"crypto/md5"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

var (
	errEof             = errors.New("eof")
	errNotConnected    = errors.New("not connected")
	errExpectingAccept = errors.New("expecting accept")
)

func readOne(scanner *bufio.Scanner) (string, error) {
	if !scanner.Scan() {
		return "", errEof
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return scanner.Text(), nil
}

type EventghostConnection struct {
	secret    string
	addr      string
	connected bool
	outgoing  chan []byte
}

func NewEventghostConnection(addr string, secret string, logger *log.Logger) *EventghostConnection {
	egc := &EventghostConnection{
		secret:    secret,
		addr:      addr,
		connected: false,
		outgoing:  make(chan []byte, 10),
	}

	// connection manager
	go func() {
		// do forever
		for {
			if err := egc.connectAuthAndServe(); err != nil {
				egc.connected = false
				logger.Printf("failed to connect/auth: %s", err.Error())
				time.Sleep(5 * time.Second)
				continue // try again
			}
		}
	}()

	return egc
}

func (e *EventghostConnection) connectAuthAndServe() error {
	dialer := net.Dialer{
		Timeout: 3 * time.Second,
	}
	conn, err := dialer.Dial("tcp", e.addr)
	if err != nil {
		return err
	}

	if _, err := conn.Write([]byte(magicKnock + "\n")); err != nil {
		return err
	}

	lineScanner := bufio.NewScanner(conn)

	challenge, err := readOne(lineScanner)
	if err != nil {
		return err
	}

	challengeResponse := fmt.Sprintf("%X", md5.Sum([]byte(challenge+":"+e.secret)))

	if _, err := conn.Write([]byte(challengeResponse + "\n")); err != nil {
		return err
	}

	succesfullyConnected := make(chan interface{})

	// for some reason EventGhost does not respond anything (but keeps the TCP socket open)
	// if we give the wrong password (even though the source code looks like it should
	// properly close). this was observed with tcpdump
	go func() {
		select {
		case <-time.After(1 * time.Second):
			conn.Close()
			return
		case <-succesfullyConnected:
			return // cancels closing timeout
		}
	}()

	accept, errAccept := readOne(lineScanner)
	if errAccept != nil {
		close(succesfullyConnected)
		return errAccept
	}
	close(succesfullyConnected)

	if accept != "accept" {
		return errExpectingAccept
	}

	e.connected = true

	for outgoing := range e.outgoing {
		if _, err := conn.Write(outgoing); err != nil {
			return err
		}
	}

	return nil
}

func (e *EventghostConnection) Send(event string, payload []string) error {
	// here's a race condition, but the worst thing that could happen is that
	// we report out-of-date info or push an item into a buffered channel
	if !e.connected {
		return errNotConnected
	}

	msg := []string{}

	for _, payloadItem := range payload {
		msg = append(msg, "payload "+payloadItem)
	}

	msg = append(msg, event)

	msgRaw := []byte(strings.Join(msg, "\n") + "\n")

	e.outgoing <- msgRaw

	return nil
}
