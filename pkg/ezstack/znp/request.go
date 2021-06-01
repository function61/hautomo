package znp

import (
	"context"

	"github.com/function61/hautomo/pkg/ezstack/znp/unp"
)

// represents either an async or sync outgoing UNP frame
type Outgoing interface {
	Request() *unp.Frame
}

// async

type Async struct {
	request *unp.Frame
}

func NewAsync(request *unp.Frame) *Async {
	return &Async{request}
}

func (a *Async) Request() *unp.Frame {
	return a.request
}

// sync (kind of like a promise)

type Sync struct {
	ctx      context.Context
	request  *unp.Frame // always known after ctor
	response *unp.Frame // only set after doneErr signalled
	doneErr  chan error // signalled after MarkDone() call
}

func NewSync(ctx context.Context, request *unp.Frame) *Sync {
	return &Sync{ctx, request, nil, make(chan error, 1)}
}

func (s *Sync) Request() *unp.Frame {
	return s.request
}

// NOTE: this can be called only once!
func (s *Sync) MarkDone(response *unp.Frame, err error) {
	s.response = response
	s.doneErr <- err
}

// NOTE: this can be called only once!
func (s *Sync) WaitForResponse(ctx context.Context) (*unp.Frame, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-s.doneErr:
		return s.response, err
	}
}
