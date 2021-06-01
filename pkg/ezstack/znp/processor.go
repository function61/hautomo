// Zigbee Network Processor. UNP is abstraction for Zigbee/Bluetooth, ZNP is concrete Zigbee+UNP
package znp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/function61/gokit/log/logex"
	"github.com/function61/gokit/sync/taskrunner"
	"github.com/function61/hautomo/pkg/ezstack/binstruct"
	"github.com/function61/hautomo/pkg/ezstack/znp/unp"
)

type Znp struct {
	unp          *unp.Unp
	outbound     chan Outgoing
	inbound      chan *unp.Frame
	asyncInbound chan interface{}
	errors       chan error
	inFramesLog  chan *unp.Frame
	outFramesLog chan *unp.Frame
	logger       *log.Logger
}

func New(unifiedProcessor *unp.Unp, logger *log.Logger) *Znp {
	return &Znp{
		unp:          unifiedProcessor,
		outbound:     make(chan Outgoing),
		inbound:      make(chan *unp.Frame),
		asyncInbound: make(chan interface{}, 10), // capacity fixes: https://github.com/dyrkin/znp-go/issues/1
		errors:       make(chan error, 100),
		inFramesLog:  make(chan *unp.Frame, 100),
		outFramesLog: make(chan *unp.Frame, 100),
		logger:       logger,
	}
}

func (z *Znp) Errors() chan error {
	return z.errors
}

func (z *Znp) AsyncInbound() chan interface{} {
	return z.asyncInbound
}

func (z *Znp) InFramesLog() chan *unp.Frame {
	return z.inFramesLog
}

func (z *Znp) OutFramesLog() chan *unp.Frame {
	return z.outFramesLog
}

func (z *Znp) Run(ctx context.Context) error {
	// there can be at most one sync request in-flight at a time
	var inflightSyncRequest *Sync

	// we can't combine the inbound/outbound stuff because some block with long processing and it can lead to deadlock

	tasks := taskrunner.New(ctx, z.logger)

	tasks.Start("inbound", func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case frame := <-z.inbound:
				switch frame.CommandType {
				case unp.C_SRSP: // sync response
					if inflightSyncRequest == nil {
						return errors.New("received sync response but no in-flight sync request")
					}

					responseErr := func() error {
						// TODO: isn't it an error if subsystem != S_RES0 ?
						if frame.Subsystem == unp.S_RES0 && frame.Command == 0 { // *Command* not to be confused with *CommandType*
							return unp.ErrorCode(frame.Payload[0]).AsError()
						} else {
							return nil
						}
					}()

					inflightSyncRequest.MarkDone(frame, responseErr)

					// outbound loop will clear inflightSyncRequest
				case unp.C_AREQ: // async response
					// this "new"'s a concrete struct dynamically. it's now an empty struct ..
					command, err := NewConcreteAsyncCommand(frame.Subsystem, frame.Command)
					if err != nil {
						select {
						case z.errors <- fmt.Errorf("%w: %v", err, frame):
						default:
						}

						break
					}

					// .. which we'll fill dynamically here
					if err := binstruct.Decode(frame.Payload, command); err != nil {
						return err // TODO: log error?
					}

					select {
					case z.asyncInbound <- command:
					default: // TODO: should this be a more fatal error?
						logex.Levels(z.logger).Error.Println("z.asyncInbound full")
					}
				default:
					select {
					case z.errors <- fmt.Errorf("unsupported frame received type: %v ", frame):
					default:
					}
				}
			}
		}
	})

	tasks.Start("outbound", func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case outgoing := <-z.outbound:
				switch req := outgoing.(type) {
				case *Sync:
					if inflightSyncRequest != nil { // shouldn't happen
						return errors.New("trying to start new sync request while another in-flight is in progress")
					}

					inflightSyncRequest = req

					logInboundOrOutboundFrame(req.Request(), z.outFramesLog)

					if err := z.unp.WriteFrame(req.Request()); err != nil {
						req.MarkDone(nil, err)

						return err // stop ("crash") outbound task since transport to radio is broken
					}

					// don't send out next outbound frame (even async ones. TODO: is this required?)
					// until we've received response to current in-flight sync request (or it was errored in WriteFrame() )
					<-req.ctx.Done()

					inflightSyncRequest = nil
				case *Async:
					logInboundOrOutboundFrame(req.Request(), z.outFramesLog)

					if err := z.unp.WriteFrame(req.Request()); err != nil {
						return err
					}
				}
			}
		}
	})

	tasks.Start("serial RX", func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
			}

			frame, err := z.unp.ReadFrame()
			if err != nil {
				// context cancelled, so as a hack our parent probably closed serial port so we
				// can exit from blocking ReadFrame() call
				if err := ctx.Err(); err != nil {
					return nil
				}

				select {
				case z.errors <- err:
				default:
				}

				return err // stop reading frames. stop task -> stops znp (which our parent can also react to)
			}

			logInboundOrOutboundFrame(frame, z.inFramesLog)

			z.inbound <- frame
		}
	})

	return tasks.Wait()
}

func (z *Znp) SendSync(subsystem unp.Subsystem, command byte, req interface{}, resp interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	outgoing := NewSync(ctx, &unp.Frame{
		CommandType: unp.C_SREQ,
		Subsystem:   subsystem,
		Command:     command,
		Payload:     binstruct.Encode(req),
	})

	// it will be now queued for sending. if previous sync request is waiting, it will only be
	// sent after the previous one has had a response.
	z.outbound <- outgoing

	responseFrame, err := outgoing.WaitForResponse(ctx)
	if err != nil { // error can also be transport error between us and the radio (= not the end device)
		return err
	} else {
		return binstruct.Decode(responseFrame.Payload, resp)
	}
}

func (z *Znp) SendAsync(subsystem unp.Subsystem, command byte, req interface{}, resp interface{}) {
	z.outbound <- NewAsync(&unp.Frame{
		CommandType: unp.C_AREQ,
		Subsystem:   subsystem,
		Command:     command,
		Payload:     binstruct.Encode(req),
	})
}

func logInboundOrOutboundFrame(frame *unp.Frame, logger chan *unp.Frame) {
	go func() {
		select {
		case logger <- frame:
		default:
		}
	}()
}
