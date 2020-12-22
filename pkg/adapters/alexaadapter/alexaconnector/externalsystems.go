package alexaconnector

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/function61/gokit/aws/s3facade"
	"github.com/function61/gokit/ezhttp"
	"github.com/function61/hautomo/pkg/adapters/alexaadapter/aamessages"
)

// abstracts away external side effects, so core logic can be tested
type ExternalSystems interface {
	TokenToUserId(ctx context.Context, token string) (string, error)
	FetchDiscoveryFile(ctx context.Context, userId string) (io.ReadCloser, error)
	SendCommand(ctx context.Context, queue string, command aamessages.Message) error
}

// the below things are essentially the only things from this package that are left untested

type externalSystems struct {
	discoveryBucket *s3facade.BucketContext
	sqs             *sqs.SQS
}

func NewExternalSystems() (ExternalSystems, error) {
	discoveryBucket, err := s3facade.Bucket("homeautomation.function61.com", nil, "us-east-1")
	if err != nil {
		return nil, err
	}

	return &externalSystems{
		discoveryBucket: discoveryBucket,
		sqs: sqs.New(session.Must(session.NewSession()), &aws.Config{
			Region: aws.String(endpoints.UsEast1RegionID),
		}),
	}, nil
}

func (e *externalSystems) TokenToUserId(ctx context.Context, token string) (string, error) {
	resp := struct {
		UserId string `json:"user_id"`
	}{}
	if _, err := ezhttp.Get(
		ctx,
		"https://api.amazon.com/user/profile",
		ezhttp.AuthBearer(token),
		ezhttp.RespondsJson(&resp, true),
	); err != nil {
		return "", nil
	}

	return resp.UserId, nil
}

func (e *externalSystems) FetchDiscoveryFile(ctx context.Context, userId string) (io.ReadCloser, error) {
	// FIXME: temp workaround until I can update one user's (mr. V) device
	if fmt.Sprintf("%x", sha1.Sum([]byte(userId))) == "1b3206a6fd66579cbbbf1f671e3c4a9f9417314c" {
		userId = "a93be8a8c85febf938c6edd0b1dc5c8f32dccb3f"
	}

	res, err := e.discoveryBucket.S3.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: e.discoveryBucket.Name,
		Key:    aws.String(fmt.Sprintf("discovery/%s.json", userId)),
	})
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

func (e *externalSystems) SendCommand(ctx context.Context, queue string, command aamessages.Message) error {
	commandSerialized, err := aamessages.Marshal(command)
	if err != nil {
		return err
	}

	_, err = e.sqs.SendMessageWithContext(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queue),
		MessageBody: aws.String(commandSerialized),
	})
	return err
}
