package homeassistantadapter

import (
	"context"

	"github.com/function61/gokit/log/logex"
	"github.com/function61/hautomo/pkg/homeassistant"
)

func makeUrlCheckerSensor(
	entityId string,
	url string,
	ha *homeassistant.MqttClient,
	logl *logex.Leveled,
) (*homeassistant.Entity, func(ctx context.Context) error) {
	sensor := homeassistant.NewSensor(
		entityId,
		url,
		homeassistant.DeviceClassDefault,
		topicPrefix,
		false)

	changeDetector := newUrlChangeDetector(url)

	return sensor, func(ctx context.Context) error {
		// does HEAD request with caching headers to check if the resource has changed
		changed, err := changeDetector.Detect(ctx)
		if err != nil {
			return err
		}

		if !changed {
			return nil
		}

		logl.Info.Printf("%s changed", entityId)

		return ha.PublishState(sensor, cacheBust(url))
	}
}
