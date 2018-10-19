package ikeatradfri

import (
	"errors"
	"fmt"
	"os/exec"
)

// TODO: use native Golang COAP + DTLS to get rid of "$ coap-client" dependency which you manually have to compile

const (
	turnOnMsg           = `{ "3311": [{ "5850": 1 }] }`
	turnOffMsg          = `{ "3311": [{ "5850": 0 }] }`
	dimWithoutFadingMsg = `{ "3311": [{ "5851": %d }] }`
)

type CoapClient struct {
	baseUrl      string
	username     string
	preSharedKey string
}

func NewCoapClient(baseUrl string, username string, preSharedKey string) *CoapClient {
	return &CoapClient{
		baseUrl:      baseUrl,
		username:     username,
		preSharedKey: preSharedKey,
	}
}

func (c *CoapClient) Put(path string, data string) error {
	coapCmd := exec.Command(
		"coap-client",
		"-B", "2", // break operation after waiting given seconds
		"-u", c.username,
		"-k", c.preSharedKey,
		"-m", "put",
		"-e", data,
		c.baseUrl+path)

	// coap-client returns success status even if request does not succeed,
	// OS-level errors probably mean that the binary was not found
	_, err := coapCmd.CombinedOutput()
	if err != nil {
		return err
	}

	return nil

	/*
		lines := strings.Split(string(output), "\n")

		if len(lines) < 3 {
			return errors.New("unexpected response structure")
		}

		// last line is response
		responseLine := lines[len(lines)-1]

		log.Printf("Put response: %s", responseLine)

	*/
}

func TurnOn(deviceId string, client *CoapClient) error {
	return client.Put(deviceEndpoint(deviceId), turnOnMsg)
}

func TurnOff(deviceId string, client *CoapClient) error {
	return client.Put(deviceEndpoint(deviceId), turnOffMsg)
}

func DimWithoutFading(deviceId string, to int, client *CoapClient) error {
	if to < 0 || to > 254 {
		return errors.New("invalid argument")
	}

	return client.Put(
		deviceEndpoint(deviceId),
		fmt.Sprintf(dimWithoutFadingMsg, to))
}

func deviceEndpoint(deviceId string) string {
	return "/15001/" + deviceId
}
