package client

import (
	"bytes"
	"encoding/gob"
	"github.com/function61/home-automation-hub/pkg/happylights/types"
	"net"
)

func SendRequest(serverAddr string, req types.LightRequest) error {
	conn, errDial := net.Dial("udp", serverAddr+":9092")
	if errDial != nil {
		return errDial
	}
	defer conn.Close()

	var reqAsGob bytes.Buffer
	enc := gob.NewEncoder(&reqAsGob)

	errEnc := enc.Encode(req)
	if errEnc != nil {
		return errEnc
	}

	_, err := conn.Write(reqAsGob.Bytes())

	return err
}
