// Sync state from Hautomo to Home Assistant, along with support for pushing remote URL
// changes (images / RSS feeds) to Home Assistant
package homeassistantadapter

import (
	"context"
	"fmt"
	"time"

	"github.com/function61/gokit/log/logex"
	"github.com/function61/hautomo/pkg/hapitypes"
	"github.com/function61/hautomo/pkg/homeassistant"
)

var (
	topicPrefix = homeassistant.NewTopicPrefix("hautomo")
)

func Start(ctx context.Context, adapter *hapitypes.Adapter) error {
	ha, err := homeassistant.NewMqttClient(adapter.Conf.Url, adapter.Logl)
	ha, err := homeassistant.NewMqttClient(adapter.Conf.Url, "Hautomo-Home-Assistant", adapter.Logl)
	if err != nil {
		return fmt.Errorf("NewMqttClient: %w", err)
	}

	homeAssistantInboundCommand, err := ha.SubscribeForCommands(topicPrefix)
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

			switch cmd.Payload {
			case "ON":
				adapter.Receive(hapitypes.NewPowerEvent(cmd.EntityId, hapitypes.PowerKindOn, true))
			case "OFF":
				adapter.Receive(hapitypes.NewPowerEvent(cmd.EntityId, hapitypes.PowerKindOff, true))
			default:
				adapter.Logl.Error.Printf("unrecognized state: %s", cmd.Payload)
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
