package particleapi

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func Invoke(device string, command string, arg string, accessToken string) error {
	apiEndpoint := "https://api.particle.io/v1/devices/" + device + "/" + command

	urlValues := url.Values{}
	urlValues.Add("access_token", accessToken)
	urlValues.Add("arg", arg)

	req, _ := http.NewRequest("POST", apiEndpoint, strings.NewReader(urlValues.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	httpClient := http.Client{}
	response, errTransport := httpClient.Do(req)
	if errTransport != nil {
		return errTransport
	}

	if response.StatusCode < 200 || response.StatusCode > 299 {
		return fmt.Errorf("Response not 2xx: got %d", response.StatusCode)
	}

	return nil
}
