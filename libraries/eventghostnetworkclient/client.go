package eventghostnetworkclient

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

func NewEventghostConnection(addr string, secret string) *EventghostConnection {
	egc := &EventghostConnection{
		secret:    secret,
		addr:      addr,
		connected: false,
		outgoing:  make(chan []byte, 10),
	}

	// keepalive thread
	go func() {
		for {
			time.Sleep(30 * time.Second)

			if egc.connected {
				egc.Send("keepalive", nil)
			}
		}
	}()

	// connection manager
	go func() {
		// do forever
		for {
			if err := egc.connectAuthAndServe(); err != nil {
				egc.connected = false
				log.Printf("EventghostConnection: failed to connect/auth: %s", err.Error())
				time.Sleep(5 * time.Second)
				continue // try again
			}
		}
	}()

	return egc
}

func (e *EventghostConnection) connectAuthAndServe() error {
	conn, err := net.Dial("tcp", e.addr)
	if err != nil {
		return err
	}

	if _, err := conn.Write([]byte("quintessence\n")); err != nil {
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

	accept, errAccept := readOne(lineScanner)
	if errAccept != nil {
		return errAccept
	}

	if accept != "accept" {
		return errExpectingAccept
	}

	e.connected = true

	log.Printf("EventghostConnection: connected")

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
