package homeassistantadapter

import (
	"context"
	"net/http"

	"github.com/function61/gokit/net/http/ezhttp"
)

type urlChangeDetector struct {
	url      string
	lastEtag string
}

func newUrlChangeDetector(url string) *urlChangeDetector {
	return &urlChangeDetector{url, ""}
}

func (u *urlChangeDetector) Detect(ctx context.Context) (bool, error) {
	opts := []ezhttp.ConfigPiece{}

	if u.lastEtag != "" {
		opts = append(opts, ezhttp.Header("If-None-Match", u.lastEtag))
	}

	// do a HEAD request to conserve resources as much as possible
	res, err := ezhttp.Head(ctx, u.url, opts...)
	if err != nil {
		return false, err
	}
	if res.StatusCode == http.StatusNotModified {
		return false, nil // not modified
	}

	u.lastEtag = res.Header.Get("ETag")

	return true, nil // is modified
}
