package homeassistantadapter

import (
	"context"
	"fmt"

	"github.com/function61/gokit/cryptorandombytes"
)

func launch(ctx context.Context, fn func(ctx context.Context) error) <-chan error {
	ch := make(chan error)
	go func() {
		ch <- fn(ctx)
	}()
	return ch
}

func launchAndWaitMany(
	ctx context.Context,
	errFn func(error),
	tasks ...func(ctx context.Context) error,
) error {
	var firstError error

	results := []<-chan error{}

	// start all
	for _, task := range tasks {
		results = append(results, launch(ctx, task))
	}

	// wait for all their results
	for _, resultCh := range results {
		if err := <-resultCh; err != nil {
			if firstError == nil {
				firstError = err
			}

			errFn(err)
		}
	}

	return firstError
}

func cacheBust(url string) string {
	return fmt.Sprintf("%s?cachebust=%s", url, cryptorandombytes.Base64Url(4))
}
