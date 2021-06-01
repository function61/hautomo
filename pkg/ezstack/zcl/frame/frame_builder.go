package frame

import (
	"errors"

	"github.com/function61/hautomo/pkg/ezstack/binstruct"
)

type frameConfiguration struct {
	transactionIdProvider            func() uint8
	frameType                        FrameType
	frameTypeConfigured              bool
	manufacturerCode                 uint16
	manufacturerCodeConfigured       bool
	direction                        Direction
	directionConfigured              bool
	disableDefaultResponse           bool
	disableDefaultResponseConfigured bool
	commandId                        uint8
	commandIdConfigured              bool
	command                          interface{}
	commandConfigured                bool
}

type Builder interface {
	FrameType(frameType FrameType) Builder
	ManufacturerCode(manufacturerCode uint16) Builder
	Direction(direction Direction) Builder
	DisableDefaultResponse(disableDefaultResponse bool) Builder
	CommandId(commandId uint8) Builder
	Command(command interface{}) Builder
	Build() (*Frame, error)
}

var (
	defaultTransactionIdProvider = MakeDefaultTransactionIdProvider()
)

func New() Builder {
	return &frameConfiguration{transactionIdProvider: defaultTransactionIdProvider}
}

func (f *frameConfiguration) IdGenerator(transactionIdProvider func() uint8) Builder {
	f.transactionIdProvider = transactionIdProvider
	return f
}

func (f *frameConfiguration) FrameType(frameType FrameType) Builder {
	f.frameType = frameType
	f.frameTypeConfigured = true
	return f
}

func (f *frameConfiguration) ManufacturerCode(manufacturerCode uint16) Builder {
	f.manufacturerCode = manufacturerCode
	f.manufacturerCodeConfigured = true
	return f
}

func (f *frameConfiguration) Direction(direction Direction) Builder {
	f.direction = direction
	f.directionConfigured = true
	return f
}

func (f *frameConfiguration) DisableDefaultResponse(disableDefaultResponse bool) Builder {
	f.disableDefaultResponse = disableDefaultResponse
	f.disableDefaultResponseConfigured = true
	return f
}

func (f *frameConfiguration) CommandId(commandId uint8) Builder {
	f.commandId = commandId
	f.commandIdConfigured = true
	return f
}

func (f *frameConfiguration) Command(command interface{}) Builder {
	f.command = command
	f.commandConfigured = true
	return f
}

func (f *frameConfiguration) Build() (*Frame, error) {
	if err := f.validateConfiguration(); err != nil {
		return nil, err
	}

	// FIXME: why not []byte ?
	payload := func() []uint8 {
		if f.commandConfigured {
			return binstruct.Encode(f.command)
		} else {
			return []uint8{} // FIXME: why not nil?
		}
	}()

	return &Frame{
		FrameControl: &FrameControl{
			FrameType:              f.frameType,
			ManufacturerSpecific:   flag(f.manufacturerCodeConfigured),
			Direction:              f.direction,
			DisableDefaultResponse: flag(f.disableDefaultResponse),
		},
		ManufacturerCode:          f.manufacturerCode,
		TransactionSequenceNumber: f.transactionIdProvider(),
		CommandIdentifier:         f.commandId,
		Payload:                   payload,
	}, nil
}

func (f *frameConfiguration) validateConfiguration() error {
	if !f.frameTypeConfigured {
		return errors.New("frame type must be set")
	}
	if !f.commandIdConfigured {
		return errors.New("command id must be set")
	}
	if !f.directionConfigured {
		return errors.New("direction must be set")
	}
	return nil
}

func flag(flag bool) uint8 {
	if flag {
		return 1
	} else {
		return 0
	}
}

func MakeDefaultTransactionIdProvider() func() uint8 {
	transactionId := uint8(1)
	return func() uint8 {
		transactionId = transactionId + 1
		if transactionId > 255 {
			transactionId = 1
		}
		return transactionId
	}
}
