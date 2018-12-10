package alexadevicesync

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
)

type AlexaConnectorDevice struct {
	Id              string   `json:"id"`
	FriendlyName    string   `json:"friendly_name"`
	Description     string   `json:"description"`
	DisplayCategory string   `json:"display_category"`
	CapabilityCodes []string `json:"capability_codes"`
}

type AlexaConnectorSpec struct {
	Queue   string                 `json:"queue"`
	Devices []AlexaConnectorDevice `json:"devices"`
}

// https://developer.amazon.com/docs/device-apis/alexa-discovery.html#display-categories
var supportedDisplayCategories = map[string]bool{
	"LIGHT":     true,
	"SPEAKER":   true,
	"TV":        true,
	"SMARTPLUG": true,
}

var supportedCapabilities = map[string]bool{
	"PowerController":            true,
	"BrightnessController":       true,
	"ColorController":            true,
	"PlaybackController":         true,
	"ColorTemperatureController": true,
}

func Sync(conf *hapitypes.ConfigFile) error {
	var sqsAdapter *hapitypes.AdapterConfig = nil

	for _, adapter := range conf.Adapters {
		if adapter.SqsQueueUrl != "" {
			copied := adapter // doesn't work: sqsAdapter = &*adapter
			sqsAdapter = &copied
		}
	}

	if sqsAdapter == nil || sqsAdapter.SqsAlexaUsertokenHash == "" {
		return errors.New("invalid configuration for SyncToAlexaConnector")
	}

	devices := []AlexaConnectorDevice{}

	for _, device := range conf.Devices {
		if _, ok := supportedDisplayCategories[device.AlexaCategory]; !ok {
			return fmt.Errorf("unsupported AlexaCategory: %s", device.AlexaCategory)
		}

		for _, capability := range device.AlexaCapabilities {
			if _, ok := supportedCapabilities[capability]; !ok {
				return fmt.Errorf("unsupported AlexaCapability: %s", capability)
			}
		}

		devices = append(devices, AlexaConnectorDevice{
			Id:              device.DeviceId,
			FriendlyName:    device.Name,
			Description:     device.Description,
			DisplayCategory: device.AlexaCategory,
			CapabilityCodes: device.AlexaCapabilities,
		})
	}

	for _, deviceGroup := range conf.DeviceGroups {
		devices = append(devices, AlexaConnectorDevice{
			Id:              deviceGroup.Id,
			FriendlyName:    deviceGroup.Name,
			Description:     "Device group",
			DisplayCategory: "LIGHT",                     // TODO
			CapabilityCodes: []string{"PowerController"}, // TODO
		})
	}

	return uploadAlexaConnectorSpec(
		sqsAdapter.SqsAlexaUsertokenHash,
		AlexaConnectorSpec{
			Queue:   sqsAdapter.SqsQueueUrl,
			Devices: devices,
		},
		sqsAdapter.SqsKeyId,
		sqsAdapter.SqsKeySecret)
}

func uploadAlexaConnectorSpec(userTokenHash string, spec AlexaConnectorSpec, accessKeyId string, accessKeySecret string) error {
	jsonBytes, errJson := json.MarshalIndent(&spec, "", "  ")
	if errJson != nil {
		return errJson
	}

	svc := s3.New(session.Must(session.NewSession()), &aws.Config{
		Region:      aws.String(endpoints.UsEast1RegionID),
		Credentials: credentials.NewStaticCredentials(accessKeyId, accessKeySecret, ""),
	})

	_, err := svc.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String("homeautomation.function61.com"),
		Key:         aws.String("discovery/" + userTokenHash + ".json"),
		Body:        bytes.NewReader(jsonBytes),
		ContentType: aws.String("application/json"),
	})

	return err
}
