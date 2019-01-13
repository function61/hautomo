package zigbee2mqttadapter

import (
	"fmt"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
	"strings"
)

func parseMsgPayload(topicName string, message string) *hapitypes.PublishEvent {
	// "zigbee2mqtt/0x00158d000227a73c" => "0x00158d000227a73c"
	dev := topicName[len(z2mTopicPrefix):]

	if strings.Contains(message, `"click":"single"`) {
		return hapitypes.NewPublishEvent(fmt.Sprintf("zigbee2mqtt:%s:click", dev))
	}

	// {"battery":100,"voltage":3055,"linkquality":47,"click":"double"}
	if strings.Contains(message, `"click":"double"`) {
		return hapitypes.NewPublishEvent(fmt.Sprintf("zigbee2mqtt:%s:double", dev))
	}

	// {"battery":100,"voltage":3055,"linkquality":60,"contact":false}
	if strings.Contains(message, `"contact":false`) {
		return hapitypes.NewPublishEvent(fmt.Sprintf("zigbee2mqtt:%s:contact:false", dev))
	}

	if strings.Contains(message, `"contact":true`) {
		return hapitypes.NewPublishEvent(fmt.Sprintf("zigbee2mqtt:%s:contact:true", dev))
	}

	return nil
}
