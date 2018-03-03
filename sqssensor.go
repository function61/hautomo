package main

import (
	"./hapitypes"
	"./util/stopper"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"log"
	"regexp"
	"time"
)

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

func sqsPollerLoop(app *Application, queueUrl string, accessKeyId string, accessKeySecret string, stopper *stopper.Stopper) {
	defer stopper.Done()

	sess := session.Must(session.NewSession())

	sqsClient := sqs.New(sess, &aws.Config{
		Region:      aws.String(endpoints.UsEast1RegionID),
		Credentials: credentials.NewStaticCredentials(accessKeyId, accessKeySecret, ""),
	})

	log.Println("sqsPollerLoop: started")

	for {
		select {
		case <-stopper.ShouldStop:
			log.Println("sqsPollerLoop: stopping")
			return
		default:
		}

		result, receiveErr := sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
			MaxNumberOfMessages: aws.Int64(10),
			QueueUrl:            &queueUrl,
			WaitTimeSeconds:     aws.Int64(10),
		})

		if receiveErr != nil {
			log.Printf("sqsPollerLoop: error in ReceiveMessage(): %s", receiveErr.Error())
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
				log.Printf("sqsPollerLoop: parse error: " + msgParseErr.Error())
				continue
			}

			switch msgType {
			case "turn_on":
				var req TurnOnRequest
				if err := json.Unmarshal([]byte(msgJsonBody), &req); err != nil {
					panic(err)
				}

				app.powerEvent <- hapitypes.NewPowerEvent(req.DeviceIdOrDeviceGroupId, hapitypes.PowerKindOn)
			case "turn_off":
				var req TurnOffRequest
				if err := json.Unmarshal([]byte(msgJsonBody), &req); err != nil {
					panic(err)
				}

				app.powerEvent <- hapitypes.NewPowerEvent(req.DeviceIdOrDeviceGroupId, hapitypes.PowerKindOff)
			case "color":
				var req ColorRequest
				if err := json.Unmarshal([]byte(msgJsonBody), &req); err != nil {
					panic(err)
				}

				app.colorEvent <- hapitypes.NewColorMsg(req.DeviceIdOrDeviceGroupId, hapitypes.RGB{req.Red, req.Green, req.Blue})
			case "brightness":
				var req BrightnessRequest
				if err := json.Unmarshal([]byte(msgJsonBody), &req); err != nil {
					panic(err)
				}

				app.brightnessEvent <- hapitypes.NewBrightnessEvent(req.DeviceIdOrDeviceGroupId, req.Brightness)
			case "playback":
				var req PlaybackRequest
				if err := json.Unmarshal([]byte(msgJsonBody), &req); err != nil {
					panic(err)
				}

				app.playbackEvent <- hapitypes.NewPlaybackEvent(req.DeviceIdOrDeviceGroupId, req.Action)
			default:
				log.Printf("sqsPollerLoop: unknown msgType: " + msgType)
			}
		}

		if len(ackList) > 0 {
			log.Printf("sqsPollerLoop: acking %d message(s)", len(ackList))

			_, err := sqsClient.DeleteMessageBatch(&sqs.DeleteMessageBatchInput{
				Entries:  ackList,
				QueueUrl: &queueUrl,
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
