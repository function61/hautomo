package ikeatradfri

import (
	"os/exec"
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

	// coap-client returns success status even if request does not succeed (no hope from
	// looking at stdout/stderr either), OS-level errors probably mean that the binary was
	// not found
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
