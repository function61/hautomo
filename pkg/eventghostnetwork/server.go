package eventghostnetwork

import (
	"bufio"
	"io"
	"log"
	"net"
	"strings"

	"github.com/function61/gokit/stopper"
	"github.com/function61/gokit/tcpkeepalive"
)

// you can give multiple passwords to differentiate between multiple computers
func RunServer(passwords []string, events eventListener, stop *stopper.Stopper) error {
	tcpListener, err := net.Listen("tcp", ":3762")
	if err != nil {
		return err
	}

	go func() {
		defer stop.Done()
		<-stop.Signal

		if err := tcpListener.Close(); err != nil {
			log.Printf("error closing tcpListener: %v", err)
		}
	}()

	handleOneClient := func(conn net.Conn) {
		if err := serverHandleClient(passwords, events, conn); err != nil {
			log.Printf("serverHandleClient: %v", err)
		}
	}

	for {
		conn, err := tcpListener.Accept()
		if err != nil {
			return err // probably cannot Accept() anymore
		}

		go handleOneClient(conn)
	}
}

func serverHandleClient(passwords []string, listener eventListener, conn net.Conn) error {
	if err := tcpkeepalive.Enable(conn.(*net.TCPConn), tcpkeepalive.DefaultDuration); err != nil {
		return err
	}

	client := newLineScanner(conn)

	state := newServerStateMachine(passwords, listener)

	for client.Scan() {
		// bufio.ScanLines() removes trailing \r\n and \n, but not \n\r which EventGhost
		// network sender sends on magic knock, and because of that we will get a leading
		// \r on the next message
		inputStripped := strings.TrimLeft(client.Text(), "\r")

		response := state.process(inputStripped)
		if response != "" {
			if _, err := conn.Write([]byte(response + "\n")); err != nil {
				return err
			}
		}

		if state.state == serverStateDisconnected {
			if err := conn.Close(); err != nil {
				return err
			}
			return nil
		}
	}
	if err := client.Err(); err != nil {
		return err
	}

	return nil
}

func newLineScanner(r io.Reader) *bufio.Scanner {
	ls := bufio.NewScanner(r)
	ls.Split(bufio.ScanLines)
	return ls
}
