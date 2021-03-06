package alexadevicesync

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	. "github.com/function61/hautomo/pkg/builtin"
	"github.com/function61/hautomo/pkg/hapitypes"
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

func Sync(sqsAdapter hapitypes.AdapterConfig, conf *hapitypes.ConfigFile) error {
	spec, err := createAlexaConnectorSpec(sqsAdapter, conf)
	if err != nil {
		return err
	}

	return uploadAlexaConnectorSpec(
		sqsAdapter.SqsAlexaUsertokenHash,
		*spec,
		sqsAdapter.SqsKeyId,
		sqsAdapter.SqsKeySecret)
}

func createAlexaConnectorSpec(sqsAdapter hapitypes.AdapterConfig, conf *hapitypes.ConfigFile) (*AlexaConnectorSpec, error) {
	if sqsAdapter.Url == "" || sqsAdapter.SqsAlexaUsertokenHash == "" {
		return nil, errors.New("invalid configuration for SyncToAlexaConnector")
	}

	devices := []AlexaConnectorDevice{}

	for _, device := range conf.Devices {
		if !device.VoiceAssistant { // require opt-in to not expose everything by default to Alexa
			continue
		}

		deviceType, err := hapitypes.ResolveDeviceType(device.Type)
		if err != nil {
			return nil, err
		}

		deviceClass := *deviceType.Class // should always be set
		if device.DeviceClassId != "" {
			deviceClass = *hapitypes.DeviceClassById[device.DeviceClassId]
		}

		caps := deviceType.Capabilities

		alexaCapabilities := []string{}
		maybePushCap(&alexaCapabilities, caps.Power, "PowerController")
		maybePushCap(&alexaCapabilities, caps.Brightness, "BrightnessController")
		maybePushCap(&alexaCapabilities, caps.Color, "ColorController")
		maybePushCap(&alexaCapabilities, caps.ColorTemperature, "ColorTemperatureController")
		maybePushCap(&alexaCapabilities, caps.Playback, "PlaybackController")

		// Alexa doesn't have shade/cover controls, so let's use PercentageController
		maybePushCap(&alexaCapabilities, caps.CoverPosition, "PowerController")
		maybePushCap(&alexaCapabilities, caps.CoverPosition, "PercentageController")

		if len(alexaCapabilities) == 0 {
			return nil, fmt.Errorf(
				"device '%s' was opt-in to Alexa, but no suitable capabilities found",
				device.DeviceId)
		}

		description := FirstNonEmpty(device.Description, device.Name)

		// Alexa seems to silently fail whole discovery if this is not set, only telling the user of
		// discovery that no new devices were found, telling nothing about the error. also there was
		// no error logs in Alexa Skills developer console to troubleshoot this issue. wonderful stuff.
		if description == "" {
			return nil, fmt.Errorf("device '%s': description and name cannot be empty", device.DeviceId)
		}

		devices = append(devices, AlexaConnectorDevice{
			Id:              device.DeviceId,
			FriendlyName:    device.Name,
			Description:     description,
			DisplayCategory: deviceClass.AlexaCategory,
			CapabilityCodes: alexaCapabilities,
		})
	}

	return &AlexaConnectorSpec{
		Queue:   sqsAdapter.Url,
		Devices: devices,
	}, nil
}

func uploadAlexaConnectorSpec(
	userId string,
	spec AlexaConnectorSpec,
	accessKeyId string,
	accessKeySecret string,
) error {
	jsonBytes, errJson := json.MarshalIndent(&spec, "", "  ")
	if errJson != nil {
		return errJson
	}

	svc := s3.New(session.Must(session.NewSession()), &aws.Config{
		Region:      aws.String(endpoints.UsEast1RegionID),
		Credentials: credentials.NewStaticCredentials(accessKeyId, accessKeySecret, ""),
	})

	_, err := svc.PutObjectWithContext(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String("homeautomation.function61.com"),
		Key:         aws.String(fmt.Sprintf("discovery/%s.json", userId)),
		Body:        bytes.NewReader(jsonBytes),
		ContentType: aws.String("application/json"),
	})

	return err
}

func maybePushCap(ref *[]string, hasCapability bool, capStr string) {
	if hasCapability {
		*ref = append(*ref, capStr)
	}
}
