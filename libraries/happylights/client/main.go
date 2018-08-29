package client

import (
	"bytes"
	"encoding/gob"
	"github.com/function61/home-automation-hub/libraries/happylights/types"
	"net"
)

func SendRequest(serverAddr string, req types.LightRequest) error {
	conn, err := net.Dial("udp", serverAddr+":9092")
	if err != nil {
		return err
	}
	defer conn.Close()

	var network bytes.Buffer

	enc := gob.NewEncoder(&network)

	errEnc := enc.Encode(req)
	if errEnc != nil {
		return errEnc
	}

	conn.Write(network.Bytes())

	return nil
}
