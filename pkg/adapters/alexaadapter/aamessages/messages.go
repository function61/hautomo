// Messages with which the adapter and the connector communicate by
package aamessages

import (
	"encoding/json"

	"github.com/function61/hautomo/pkg/ezstack/ezhub/hubtypes"
)

// message used to send attribute change requests to Hautomo from Lambda Alexa skill
type Message struct {
	DeviceId string
	Attrs    hubtypes.Attributes
}

func Marshal(msg Message) (string, error) {
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(msgJSON), nil
}

func Unmarshal(input string) (Message, error) {
	msg := Message{}
	return msg, json.Unmarshal([]byte(input), &msg)
}
