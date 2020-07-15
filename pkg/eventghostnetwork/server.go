package eventghostnetwork

import (
	"bufio"
	"context"
	"io"
	"log"
	"net"
	"strings"
)

// you can give multiple passwords to differentiate between multiple computers
func RunServer(ctx context.Context, passwords []string, events eventListener) error {
	tcpListener, err := net.Listen("tcp", ":3762")
	if err != nil {
		return err
	}

	handleOneClient := func(conn net.Conn) {
		if err := serverHandleClient(passwords, events, conn); err != nil {
			log.Printf("serverHandleClient: %v", err)
		}
	}

	go func() {
		<-ctx.Done()

		if err := tcpListener.Close(); err != nil {
			log.Printf("error closing tcpListener: %v", err)
		}
	}()

	for {
		conn, err := tcpListener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil // error was expected, since we got cancelled
			default:
				return err // probably cannot Accept() anymore
			}
		}

		go handleOneClient(conn)
	}
}

func serverHandleClient(passwords []string, listener eventListener, conn net.Conn) error {
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
