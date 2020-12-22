// Client for screen-server
package screenserverclient

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"strings"

	"github.com/function61/gokit/ezhttp"
)

// TODO: extract as a client to screen-server repo

type server struct {
	baseUrl string
}

func Server(baseUrl string) server {
	return server{baseUrl}
}

type screen struct {
	id string

	baseUrl string
}

func (s server) Screen(id string) screen {
	return screen{
		id:      id,
		baseUrl: s.baseUrl,
	}
}

func (s *screen) DisplayNotification(ctx context.Context, message string) error {
	endpoint := fmt.Sprintf(
		"%s/api/screen/%s/osd/notify",
		s.baseUrl,
		s.id)

	data := url.Values{}
	data.Set("msg", message)

	// TODO: form data helper to ezhttp
	_, err := ezhttp.Post(
		ctx,
		endpoint,
		ezhttp.SendBody(strings.NewReader(data.Encode()), "application/x-www-form-urlencoded"))
	return err

}

type clientClient struct {
	ip string
}

// screen-server has clients. this is a client for clients
func ClientClient(ip string) clientClient {
	return clientClient{
		ip: ip,
	}
}

// on/off work by instructing the Android-based server to acquire/release a wake lock
func (c clientClient) ScreenPowerOn(ctx context.Context) error {
	return clientClientCommand(ctx, c.ip, "wl.on")
}

func (c clientClient) ScreenPowerOff(ctx context.Context) error {
	return clientClientCommand(ctx, c.ip, "wl.off")
}

// sample URL: http://192.168.1.105:5902/sample.mp3
func (c clientClient) PlayAudio(ctx context.Context, audioUrl string) error {
	return clientClientCommand(ctx, c.ip, "audio.play", audioUrl)
}

func clientClientCommand(ctx context.Context, ip string, cmdParts ...string) error {
	conn, err := net.Dial("tcp", ip+":53000")
	if err != nil {
		return err
	}
	defer conn.Close()

	// "screen-server client success <ACTION>"
	expectBack := fmt.Sprintf("sscs\n%s\n", cmdParts[0])

	packet := fmt.Sprintf("%s\n", strings.Join(cmdParts, ","))

	if _, err = conn.Write([]byte(packet)); err != nil {
		return err
	}

	gotBack, err := ioutil.ReadAll(conn)
	if err != nil {
		return err
	}

	if expectBack != string(gotBack) {
		return fmt.Errorf("reply expectation failed, got back: %s", gotBack)
	}

	return nil
}
