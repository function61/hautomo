// Receives commands from Alexa voice assistant ("Alexa, turn off <device name>") via AWS
// SQS (so the Lambda fn doesn't need direct connection to Hautomo). The commands are
// one-way, so Alexa won't know if the command succeeded.
package alexaadapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/function61/gokit/sync/taskrunner"
	"github.com/function61/hautomo/pkg/adapters/alexaadapter/aamessages"
	"github.com/function61/hautomo/pkg/adapters/alexaadapter/alexadevicesync"
	"github.com/function61/hautomo/pkg/hapitypes"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/amazon"
)

// Start the event receiver side. Take serialized commands from SQS and translate them
// into explicit commands for Hautomo to handle
func Start(ctx context.Context, adapter *hapitypes.Adapter) error {
	oauth2AppConfig := &oauth2.Config{
		ClientID:     adapter.Conf.AlexaOauth2ClientId,
		ClientSecret: adapter.Conf.AlexaOauth2ClientSecret,

		Endpoint: amazon.Endpoint,
	}

	if oauth2AppConfig.ClientID == "" || oauth2AppConfig.ClientSecret == "" {
		return errors.New("AlexaOauth2ClientId or AlexaOauth2ClientSecret empty")
	}

	alexaUserToken, err := func() (*oauth2.Token, error) {
		if adapter.Conf.AlexaOauth2UserToken == "" {
			return nil, errors.New("empty AlexaOauth2UserToken")
		}

		tok := &oauth2.Token{}
		return tok, json.Unmarshal([]byte(adapter.Conf.AlexaOauth2UserToken), tok)
	}()
	if err != nil {
		return err
	}

	// at the start, sync our (Alexa-compatible) device registry into the connector running
	// in Lambda, so we can receive commands for them
	if adapter.Conf.SqsAlexaUsertokenHash != "" {
		if err := alexadevicesync.Sync(adapter.Conf, adapter.GetConfigFileDeprecated()); err != nil {
			return fmt.Errorf("alexadevicesync: %w", err)
		}
	}

	subTasks := taskrunner.New(ctx, adapter.Log)
	subTasks.Start("sqs-poller", func(ctx context.Context) error {
		if adapter.Conf.Url == "" {
			<-ctx.Done()
			return nil
		}

		sqsClient := sqs.New(session.Must(session.NewSession()), &aws.Config{
			Region: aws.String(endpoints.UsEast1RegionID),
			Credentials: credentials.NewStaticCredentials(
				adapter.Conf.SqsKeyId,
				adapter.Conf.SqsKeySecret,
				""),
		})

		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				// only returns stop-worthy errors, and logs non-severe ones
				if err := runOnce(ctx, sqsClient, adapter); err != nil {
					return err // stop ("crash")
				}
			}
		}
	})

	// we make an assumption that we only control one Alexa user's account
	alexaUserClient := oauth2AppConfig.Client(context.TODO(), alexaUserToken)

	for {
		select {
		case <-ctx.Done():
			return subTasks.Wait()
		case err := <-subTasks.Done(): // subtask crash
			return err // stop ("crash")
		case genericEvent := <-adapter.Outbound:
			switch e := genericEvent.(type) {
			case *hapitypes.NotificationEvent:
				// convert into contact sensor event. this is an ugly hack, because contact
				// and movement sensors are only event sources we can base routine triggers in
				// ATM.
				if err := sendContactSensorEvent(ctx, e.Device, false, alexaUserClient); err != nil {
					adapter.Logl.Error.Printf("sendContactSensorEvent: %v", err)
				}
			default:
				adapter.LogUnsupportedEvent(genericEvent)
			}
		}
	}
}

func runOnce(ctx context.Context, sqsClient *sqs.SQS, adapter *hapitypes.Adapter) error {
	result, receiveErr := sqsClient.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
		MaxNumberOfMessages: aws.Int64(10),
		QueueUrl:            &adapter.Conf.Url,
		WaitTimeSeconds:     aws.Int64(10), // use long-polling
	})

	if receiveErr != nil {
		adapter.Logl.Error.Printf("ReceiveMessage(): %s", receiveErr.Error())
		time.Sleep(5 * time.Second)
		return nil
	}

	ackList := []*sqs.DeleteMessageBatchRequestEntry{}

	for _, msg := range result.Messages {
		ackList = append(ackList, &sqs.DeleteMessageBatchRequestEntry{
			Id:            msg.MessageId,
			ReceiptHandle: msg.ReceiptHandle,
		})

		msg, errMsgParse := aamessages.Unmarshal(*msg.Body)
		if errMsgParse != nil {
			adapter.Logl.Error.Printf("aamessages.Unmarshal: %s", errMsgParse.Error())
			continue
		}

		attrs := msg.Attrs // shorthand

		if attrs.On != nil {
				adapter.Receive(hapitypes.NewPowerEvent(
					msg.DeviceId,
					func() hapitypes.PowerKind {
						if attrs.On.Value {
							return hapitypes.PowerKindOn
						} else {
							return hapitypes.PowerKindOff
						}
					}(),
					true))
		}

		if attrs.Color != nil {
			color := attrs.Color // shorthand

			adapter.Receive(hapitypes.NewColorMsg(
				msg.DeviceId,
				hapitypes.NewRGB(color.Red, color.Green, color.Blue)))
		}

		if attrs.Brightness != nil {
			adapter.Receive(hapitypes.NewBrightnessEvent(
				msg.DeviceId,
				uint(attrs.Brightness.Value)))
		}

		if attrs.ColorTemperature != nil {
			adapter.Receive(hapitypes.NewColorTemperatureEvent(
				msg.DeviceId,
				uint(attrs.ColorTemperature.Value)))
		}

		if attrs.PlaybackControl != nil {
			adapter.Receive(hapitypes.NewPlaybackEvent(
				msg.DeviceId,
				attrs.PlaybackControl.Control))
		}
	}

	if len(ackList) > 0 {
		// intentionally background context, so we won't cancel this important operation
		// if we happen to get a stop
		_, err := sqsClient.DeleteMessageBatchWithContext(context.Background(), &sqs.DeleteMessageBatchInput{
			Entries:  ackList,
			QueueUrl: &adapter.Conf.Url,
		})

		if err != nil {
			return err
		}
	}

	return nil
}
