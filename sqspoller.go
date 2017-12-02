package main

import (
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"log"
	"regexp"
	"time"
)

type TurnOnRequest struct {
	DeviceId string `json:"id"`
}

type TurnOffRequest struct {
	DeviceId string `json:"id"`
}

func sqsPollerLoop(app *Application, stopper *Stopper) {
	defer stopper.Done()

	sess := session.Must(session.NewSession())

	// access keys provided from command line
	queueUrl := "https://sqs.us-east-1.amazonaws.com/329074924855/JoonasHomeAutomation"

	sqsClient := sqs.New(sess, &aws.Config{
		Region: aws.String(endpoints.UsEast1RegionID),
	})

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
			log.Printf("MessageId %s Body = %s", *msg.MessageId, *msg.Body)

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

				_ = app.TurnOn(req.DeviceId)
			case "turn_off":
				var req TurnOffRequest
				if err := json.Unmarshal([]byte(msgJsonBody), &req); err != nil {
					panic(err)
				}

				_ = app.TurnOff(req.DeviceId)
			default:
				log.Printf("sqsPollerLoop: unknown msgType: " + msgType)
			}
		}

		if len(ackList) > 0 {
			log.Printf("deleting %d message(s)", len(ackList))

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
