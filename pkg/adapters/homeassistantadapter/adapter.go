// Sync state from Hautomo to Home Assistant, along with support for pushing remote URL
// changes (images / RSS feeds) to Home Assistant
package homeassistantadapter

import (
	"context"
	"fmt"
	"time"

	"github.com/function61/gokit/logex"
	"github.com/function61/hautomo/pkg/hapitypes"
	"github.com/function61/hautomo/pkg/homeassistant"
)

func Start(ctx context.Context, adapter *hapitypes.Adapter) error {
	ha, err := homeassistant.NewMqttClient(adapter.Conf.Url, func(task func(context.Context) error) {
		if task != nil {
			panic("not implemented")
		}
		// do nothing
	}, adapter.Logl)
	if err != nil {
		return fmt.Errorf("NewMqttClient: %w", err)
	}

	homeAssistantInboundCommand, err := ha.SubscribeForStateChanges()
	if err != nil {
		return err
	}

	entityById := map[string]*homeassistant.Entity{}

	allEntities := []*homeassistant.Entity{}
	for _, dev := range adapter.GetConfigFileDeprecated().Devices {
		typ, err := hapitypes.ResolveDeviceType(dev.Type)
		if err != nil {
			return err
		}

		if !typ.Capabilities.VirtualSwitch {
			continue
		}

		switchEntity := homeassistant.NewSwitch(dev.AdaptersDeviceId, dev.Name)
		entityById[switchEntity.Id] = switchEntity
		allEntities = append(allEntities, switchEntity)
	}

	pollingTasks := []func(context.Context) error{}

	for _, feed := range adapter.Conf.RssFeeds {
		feedSensor, feedPollerTask := makeRssFeedSensor(feed.Id, feed.Url, ha, adapter.Logl)

		allEntities = append(allEntities, feedSensor)
		pollingTasks = append(pollingTasks, feedPollerTask)
	}

	for _, urlChangeDetector := range adapter.Conf.UrlChangeDetectors {
		sensor, task := makeUrlCheckerSensor(
			urlChangeDetector.Id,
			urlChangeDetector.Url,
			ha,
			adapter.Logl)

		allEntities = append(allEntities, sensor)
		pollingTasks = append(pollingTasks, task)
	}

	if err := ha.AutodiscoverEntities(allEntities...); err != nil {
		return err
	}

	runPollingTasks := func() {
		_ = launchAndWaitMany(ctx, func(err error) {
			adapter.Logl.Error.Println(err)
		}, pollingTasks...)
	}

	// initial sync
	runPollingTasks()

	pollInterval := time.NewTicker(1 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return nil
		case cmd := <-homeAssistantInboundCommand:
			entity, found := entityById[cmd.EntityId]
			if !found {
				return fmt.Errorf("entityById: not found: %s", cmd.EntityId)
			}

			// TODO: entity id is wrong
			switch cmd.State {
			case "ON":
				adapter.Receive(hapitypes.NewPowerEvent(cmd.EntityId, hapitypes.PowerKindOn, true))
			case "OFF":
				adapter.Receive(hapitypes.NewPowerEvent(cmd.EntityId, hapitypes.PowerKindOff, true))
			default:
				adapter.Logl.Error.Printf("unrecognized state: %s", cmd.State)
			}

			// immediately send state back
			// TODO: don't do this from here
			if err := ha.PublishState(entity, cmd.State); err != nil {
				adapter.Logl.Error.Printf("PublishState: %v", err)
			}
		case <-pollInterval.C:
			runPollingTasks()
		}
	}
}

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
