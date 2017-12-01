package main

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func particleRequest(device string, command string, arg string, accessToken string) error {
	apiEndpoint := "https://api.particle.io/v1/devices/" + device + "/" + command

	urlValues := url.Values{}
	urlValues.Add("access_token", accessToken)
	urlValues.Add("arg", arg)

	req, _ := http.NewRequest("POST", apiEndpoint, strings.NewReader(urlValues.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	httpClient := http.Client{}
	_, err := httpClient.Do(req)

	return err
}

func getParticleAccessToken() (string, error) {
	accessToken := os.Getenv("PARTICLE_ACCESS_TOKEN")
	if accessToken == "" {
		return "", errors.New("getParticleAccessToken(): token not defined")
	}

	return accessToken, nil
}
