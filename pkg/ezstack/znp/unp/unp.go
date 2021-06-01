// UNP is a framing protocol (commandtype, subsystem, command, payload) for communicating with Texas Instruments radio devices
package unp

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/dyrkin/composer"
)

type CommandType byte

const (
	C_POLL CommandType = iota
	C_SREQ
	C_AREQ
	C_SRSP
	C_RES0
	C_RES1
	C_RES2
	C_RES3
)

type Subsystem byte

const (
	S_RES0 Subsystem = iota
	S_SYS
	S_MAC
	S_NWK
	S_AF
	S_ZDO
	S_SAPI
	S_UTIL
	S_DBG
	S_APP
	S_OTA
	S_ZNP
	S_SPARE_12
	S_UBL
	S_RES14
	S_APP_CNF
	S_RES16
	S_PROTOBUF
	S_RES18 // RPC_SYS_PB_NWK_MGR
	S_RES19 // RPC_SYS_PB_GW
	S_RES20 // RPC_SYS_PB_OTA_MGR
	S_GP
	S_MAX
)

type Unp struct {
	payloadLength8bits bool
	Transceiver        io.ReadWriter
	incoming           chan byte
	errors             chan error
}

type Frame struct {
	CommandType CommandType
	Subsystem   Subsystem
	Command     byte
	Payload     []byte
}

const startOfFrame byte = 0xFE

func NewWith8BitsPayloadLength(transmitter io.ReadWriter) *Unp {
	return newInternal(true, transmitter)
}

func newInternal(payloadLength8bits bool, transmitter io.ReadWriter) *Unp {
	u := &Unp{payloadLength8bits, transmitter, make(chan byte), make(chan error)}
	go u.receiver()
	return u
}

func (u *Unp) WriteFrame(frame *Frame) error {
	rendered := u.RenderFrame(frame)
	_, err := u.Transceiver.Write(rendered)
	return err
}

func (u *Unp) RenderFrame(frame *Frame) []byte {
	cmp := composer.New()
	cmd0 := ((byte(frame.CommandType << 5)) & 0xE0) | (byte(frame.Subsystem) & 0x1F)
	cmd1 := frame.Command
	cmp.Byte(startOfFrame)
	payloadLength := len(frame.Payload)
	if u.payloadLength8bits {
		cmp.Uint8(uint8(payloadLength))
	} else {
		cmp.Uint16be(uint16(payloadLength))
	}
	cmp.Byte(cmd0).Byte(cmd1).Bytes(frame.Payload)
	fcs := checksum(cmp.Make()[1:])
	cmp.Byte(fcs)
	return cmp.Make()
}

func (u *Unp) ReadFrame() (frame *Frame, err error) {
	var b byte
	var checksumBuffer bytes.Buffer

	var read = func() {
		select {
		case b = <-u.incoming:
			checksumBuffer.WriteByte(b)
		case err = <-u.errors:
		}
	}
	if read(); err != nil {
		return
	}
	if b != startOfFrame {
		// this seems to occurr if the coordinator has been running for some time and no-one has
		// consumed its output. maybe there's a ring buffer and thus once we continue consuming after
		// long pause we don't start reading from valid frame boundary?
		return nil, errors.New("Invalid start of frame")
	}
	if read(); err != nil {
		return
	}

	var payloadLength uint16
	if u.payloadLength8bits {
		payloadLength = uint16(b)
	} else {
		b1 := uint16(b) << 8
		if read(); err != nil {
			return
		}
		payloadLength = b1 | uint16(b)
	}
	if read(); err != nil {
		return
	}
	cmd0 := b
	if read(); err != nil {
		return
	}
	cmd1 := b
	payload := make([]byte, payloadLength)
	for i := 0; i < int(payloadLength); i++ {
		if read(); err != nil {
			return
		}
		payload[i] = b
	}
	checksumBytes := checksumBuffer.Bytes()
	if read(); err != nil {
		return
	}
	fcs := b
	csum := checksum(checksumBytes[1:])
	if fcs != csum {
		err = fmt.Errorf("Invalid checksum. Expected: %b, actual: %b", fcs, csum)
		return
	}
	commandType := (cmd0 & 0xE0) >> 5
	subsystem := cmd0 & 0x1F
	return &Frame{CommandType(commandType), Subsystem(subsystem), cmd1, payload}, nil
}

func checksum(buf []byte) byte {
	fcs := byte(0)
	for i := 0; i < len(buf); i++ {
		fcs ^= buf[i]
	}
	return fcs
}

func (u *Unp) receiver() {
	var buf [1]byte
	for {
		n, err := io.ReadFull(u.Transceiver, buf[:])
		if n > 0 {
			u.incoming <- buf[0]
		} else if err != io.EOF {
			u.errors <- err
		}
	}

}

// values documented in https://dev.ti.com/tirex/content/simplelink_cc13x2_sdk_2_30_00_45/docs/zstack/html/zigbee/znp_interface.html
type ErrorCode uint8

func (e ErrorCode) AsError() error {
	if err, found := errorCodeToError[uint8(e)]; found {
		return err
	} else {
		return fmt.Errorf("unknown error code: %d", e)
	}
}

var errorCodeToError = map[uint8]error{
	1: errors.New("Invalid subsystem"),
	2: errors.New("Invalid command ID"),
	3: errors.New("Invalid parameter"),
	4: errors.New("Invalid length"),
}
