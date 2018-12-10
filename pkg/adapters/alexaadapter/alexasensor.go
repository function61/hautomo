package alexaadapter

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/function61/gokit/logger"
	"github.com/function61/gokit/stopper"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
	"github.com/function61/home-automation-hub/pkg/signalfabric"
	"regexp"
	"time"
)

var log = logger.New("alexa")

type TurnOnRequest struct {
	DeviceIdOrDeviceGroupId string `json:"id"`
}

type TurnOffRequest struct {
	DeviceIdOrDeviceGroupId string `json:"id"`
}

type ColorRequest struct {
	DeviceIdOrDeviceGroupId string `json:"id"`
	Red                     uint8  `json:"red"`
	Green                   uint8  `json:"green"`
	Blue                    uint8  `json:"blue"`
}

type BrightnessRequest struct {
	DeviceIdOrDeviceGroupId string `json:"id"`
	Brightness              uint   `json:"brightness"` // 0-100
}

type PlaybackRequest struct {
	DeviceIdOrDeviceGroupId string `json:"id"`
	Action                  string `json:"action"`
}

type ColorTemperatureRequest struct {
	DeviceIdOrDeviceGroupId  string `json:"id"`
	ColorTemperatureInKelvin uint   `json:"colorTemperatureInKelvin"`
}

func StartSensor(fabric *signalfabric.Fabric, adapterConf hapitypes.AdapterConfig, stop *stopper.Stopper) {
	defer stop.Done()

	sess := session.Must(session.NewSession())

	sqsClient := sqs.New(sess, &aws.Config{
		Region: aws.String(endpoints.UsEast1RegionID),
		Credentials: credentials.NewStaticCredentials(
			adapterConf.SqsKeyId,
			adapterConf.SqsKeySecret,
			""),
	})

	log.Info("started")
	defer log.Info("stopped")

	for {
		select {
		case <-stop.Signal:
			return
		default:
		}

		result, receiveErr := sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
			MaxNumberOfMessages: aws.Int64(10),
			QueueUrl:            &adapterConf.SqsQueueUrl,
			WaitTimeSeconds:     aws.Int64(10),
		})

		if receiveErr != nil {
			log.Error(fmt.Sprintf("ReceiveMessage(): %s", receiveErr.Error()))
			time.Sleep(5 * time.Second)
			continue
		}

		ackList := []*sqs.DeleteMessageBatchRequestEntry{}

		for _, msg := range result.Messages {
			ackList = append(ackList, &sqs.DeleteMessageBatchRequestEntry{
				Id:            msg.MessageId,
				ReceiptHandle: msg.ReceiptHandle,
			})

			msgParseErr, msgType, msgJsonBody := parseMessage(*msg.Body)
			if msgParseErr != nil {
				log.Error(fmt.Sprintf("parseMessage: %s", msgParseErr.Error()))
				continue
			}

			switch msgType {
			case "turn_on":
				var req TurnOnRequest
				if err := json.Unmarshal([]byte(msgJsonBody), &req); err != nil {
					panic(err)
				}

				fabric.PowerEvent <- hapitypes.NewPowerEvent(req.DeviceIdOrDeviceGroupId, hapitypes.PowerKindOn)
			case "turn_off":
				var req TurnOffRequest
				if err := json.Unmarshal([]byte(msgJsonBody), &req); err != nil {
					panic(err)
				}

				fabric.PowerEvent <- hapitypes.NewPowerEvent(req.DeviceIdOrDeviceGroupId, hapitypes.PowerKindOff)
			case "color":
				var req ColorRequest
				if err := json.Unmarshal([]byte(msgJsonBody), &req); err != nil {
					panic(err)
				}

				fabric.ColorEvent <- hapitypes.NewColorMsg(req.DeviceIdOrDeviceGroupId, hapitypes.RGB{Red: req.Red, Green: req.Green, Blue: req.Blue})
			case "brightness":
				var req BrightnessRequest
				if err := json.Unmarshal([]byte(msgJsonBody), &req); err != nil {
					panic(err)
				}

				fabric.BrightnessEvent <- hapitypes.NewBrightnessEvent(req.DeviceIdOrDeviceGroupId, req.Brightness)
			case "playback":
				var req PlaybackRequest
				if err := json.Unmarshal([]byte(msgJsonBody), &req); err != nil {
					panic(err)
				}

				fabric.PlaybackEvent <- hapitypes.NewPlaybackEvent(req.DeviceIdOrDeviceGroupId, req.Action)
			case "colorTemperature":
				var req ColorTemperatureRequest
				if err := json.Unmarshal([]byte(msgJsonBody), &req); err != nil {
					panic(err)
				}

				fabric.ColorTemperatureEvent <- hapitypes.NewColorTemperatureEvent(
					req.DeviceIdOrDeviceGroupId,
					req.ColorTemperatureInKelvin)
			default:
				log.Error("unknown msgType: " + msgType)
			}
		}

		if len(ackList) > 0 {
			log.Debug(fmt.Sprintf("acking %d message(s)", len(ackList)))

			_, err := sqsClient.DeleteMessageBatch(&sqs.DeleteMessageBatchInput{
				Entries:  ackList,
				QueueUrl: &adapterConf.SqsQueueUrl,
			})

			if err != nil {
				panic(err)
			}
		}
	}
}

var parseMessageRegexp = regexp.MustCompile(`^([a-zA-Z_0-9]+) (.+)$`)

func parseMessage(input string) (error, string, string) {
	match := parseMessageRegexp.FindStringSubmatch(input)
	if match == nil {
		return errors.New("parseMessage(): invalid format"), "", ""
	}

	return nil, match[1], match[2]
}
