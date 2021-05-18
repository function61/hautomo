package homeassistant

import (
	"context"
	"fmt"

	"github.com/function61/gokit/net/http/ezhttp"
)

type Client struct {
	baseUrl   string
	authToken string // for non-MQTT requests
}

func NewClient(baseUrl string, authToken string) *Client {
	return &Client{baseUrl, authToken}
}

func (c *Client) TextToSpeechGetUrl(ctx context.Context, message string) (string, error) {
	req := struct {
		Platform string `json:"platform"`
		Message  string `json:"message"`
	}{
		Platform: "google_translate", // I don't know why we've to hardcode this (instead of opting for some kind of default)
		Message:  message,
	}

	res := struct {
		Url string `json:"url"`
	}{}

	if _, err := ezhttp.Post(
		ctx,
		c.baseUrl+"/api/tts_get_url",
		ezhttp.AuthBearer(c.authToken),
		ezhttp.SendJson(&req),
		ezhttp.RespondsJson(&res, false),
	); err != nil {
		return "", fmt.Errorf("TextToSpeechGetUrl: %w", err)
	}

	return res.Url, nil
}
