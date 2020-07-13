package particleapi

import (
	"context"
	"net/url"
	"strings"

	"github.com/function61/gokit/ezhttp"
)

func Invoke(device string, command string, arg string, accessToken string) error {
	urlValues := url.Values{}
	urlValues.Add("access_token", accessToken)
	urlValues.Add("arg", arg)

	ctx, cancel := context.WithTimeout(context.TODO(), ezhttp.DefaultTimeout10s)
	defer cancel()

	_, err := ezhttp.Post(
		ctx,
		"https://api.particle.io/v1/devices/"+device+"/"+command,
		ezhttp.SendBody(strings.NewReader(urlValues.Encode()), "application/x-www-form-urlencoded"))

	return err
}
