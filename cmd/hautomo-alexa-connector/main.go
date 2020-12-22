package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/function61/gokit/osutil"
	"github.com/function61/hautomo/pkg/adapters/alexaadapter/alexaconnector"
)

func main() {
	extSys, err := alexaconnector.NewExternalSystems()
	osutil.ExitIfError(err)
	handlers := alexaconnector.New(extSys, nil, nil)

	lambda.StartHandler(newBytesInAndOutHandler(func(ctx context.Context, payload []byte) ([]byte, error) {
		log.Printf("input %s", payload)

		msg, err := handlers.Handle(ctx, payload)
		if err != nil {
			return nil, fmt.Errorf("handlers.Handle: %w", err)
		}

		jsonOut, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}

		log.Printf("output %s", jsonOut)

		return jsonOut, nil
	}))
}

// TODO: is this in gokit?
type bytesInAndOutHandler struct {
	fn func(ctx context.Context, payload []byte) ([]byte, error)
}

func newBytesInAndOutHandler(fn func(ctx context.Context, payload []byte) ([]byte, error)) lambda.Handler {
	return &bytesInAndOutHandler{fn}
}

func (b *bytesInAndOutHandler) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	return b.fn(ctx, payload)
}
