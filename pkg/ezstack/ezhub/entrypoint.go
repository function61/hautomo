package ezhub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/function61/gokit/encoding/jsonfile"
	"github.com/function61/gokit/log/logex"
	"github.com/function61/gokit/net/http/httputils"
	"github.com/function61/gokit/sync/taskrunner"
	"github.com/function61/hautomo/pkg/ezstack"
	"github.com/function61/hautomo/pkg/ezstack/ezhub/deviceadapters"
	"github.com/function61/hautomo/pkg/ezstack/ezhub/homeassistantmqtt"
	"github.com/function61/hautomo/pkg/ezstack/ezhub/hubtypes"
	"github.com/function61/hautomo/pkg/ezstack/zcl/cluster"
	"github.com/function61/hautomo/pkg/ezstack/zigbee"
	"github.com/function61/hautomo/pkg/homeassistant"
)

// runs ezhub. returns when asked to stop or earlier if encountered a fatal error
func Run(
	ctx context.Context,
	joinEnable bool,
	packetCaptureFile string,
	settingsFlash bool,
	rootLogger *log.Logger,
) error {
	logl := logex.Levels(rootLogger)

	mqttPublish := make(chan homeassistantmqtt.Message, 100)
	mqttInbound := make(chan homeassistantmqtt.InboundMessage, 100)

	conf := Config{}
	if err := jsonfile.ReadDisallowUnknownFields("ezhub-config.json", &conf); err != nil {
		return err
	}

	if err := conf.Valid(); err != nil {
		return err
	}

	nodeDatabase, err := loadNodeDatabaseOrInitIfNotFound()
	if err != nil {
		return fmt.Errorf("loadNodeDatabaseOrInitIfNotFound: %w", err)
	}

	if err := homeAssistantAutoDiscovery(conf.MQTT.Addr, conf.MQTT.Prefix, nodeDatabase, rootLogger); err != nil {
		logl.Error.Printf("homeAssistantAutoDiscovery: %w", err)
	}

	stack := ezstack.New(conf.Coordinator, nodeDatabase)

	tasks := taskrunner.New(ctx, rootLogger)

	tasks.Start("mqtt-connection-loop", func(ctx context.Context) error {
		for {
			err := homeassistantmqtt.ConnectAndServe(
				ctx,
				conf.MQTT.Addr,
				conf.MQTT.Prefix,
				mqttPublish,
				mqttInbound)

			select {
			case <-ctx.Done():
				return nil
			default:
				logl.Error.Printf("mqtt-connection-loop: reconnecting due to: %v", err)
				time.Sleep(1 * time.Second)
			}
		}
	})

	tasks.Start("ezstack", func(ctx context.Context) error {
		return stack.Run(ctx, joinEnable, packetCaptureFile, settingsFlash)
	})

	if conf.HttpAddr != "" {
		srv := &http.Server{
			Addr:    conf.HttpAddr,
			Handler: createHttpApi(stack, nodeDatabase),
		}

		tasks.Start("http "+srv.Addr, func(ctx context.Context) error {
			return httputils.CancelableServer(ctx, srv, func() error { return srv.ListenAndServe() })
		})
	}

	tasks.Start("messagehandler", func(ctx context.Context) error {
		chans := zigbee.Channels()

		for {
			select {
			case <-ctx.Done():
				return nil
			case msg := <-mqttInbound: // messages TO Zigbee network
				msgJson, _ := json.Marshal(msg.Message)
				logl.Debug.Printf("MQTT inbound %s: %s", msg.DeviceId, msgJson)

				go func() {
					if err := processMQTTInboundMessage(msg, stack, nodeDatabase, mqttPublish, conf.MQTT.Prefix); err != nil {
						logl.Error.Printf("processMQTTInboundMessage: %v", err.Error())
					}
				}()
			case deviceIncomingMessage := <-chans.OnDeviceIncomingMessage(): // messages FROM Zigbee network
				wdev := nodeDatabase.GetWrappedDevice(deviceIncomingMessage.Device.IEEEAddress)
				if wdev == nil {
					return fmt.Errorf("incoming message from unknown device: %s", deviceIncomingMessage.Device.IEEEAddress)
				}

				// after this wrapper has returned, we will have informed subscribers over MQTT
				// about the changed attributes
				if err := updateAttributesAndNotifyMQTT(
					wdev,
					mqttPublish,
					conf.MQTT.Prefix,
					deviceIncomingMessage.IncomingMessage.SrcEndpoint,
					func(actx *hubtypes.AttrsCtx) error {
						// link quality is present in each frame
						wdev.State.LinkQuality = actx.Int(
							int64(deviceIncomingMessage.IncomingMessage.LinkQuality))

						return deviceadapters.ZclIncomingMessageToAttributes(deviceIncomingMessage.IncomingMessage, actx, wdev.ZigbeeDevice)
					},
				); err != nil {
					logl.Error.Printf("OnDeviceIncomingMessage: %v", err)
				}
			case _ = <-chans.OnDeviceRegistered():
				logl.Info.Println("device registered")
			case _ = <-chans.OnDeviceBecameAvailable(): // TODO: diff between registered & available?
				logl.Info.Println("device available")
			case _ = <-chans.OnDeviceUnregistered():
				logl.Info.Println("device unregistered")
			}
		}
	})

	tasks.Start("state-snapshot", createStateSnapshotTask(nodeDatabase))

	return tasks.Wait()
}

// some external system wants to control a Zigbee device
func processMQTTInboundMessage(
	inboundMsg homeassistantmqtt.InboundMessage,
	zigbee *ezstack.Stack,
	nodeDatabase *nodeDb,
	mqttPublish chan<- homeassistantmqtt.Message,
	mqttPrefix string,
) error {
	dev := nodeDatabase.GetWrappedDevice(inboundMsg.DeviceId)
	if dev == nil {
		return fmt.Errorf("device not found: %s", inboundMsg.DeviceId)
	}

	endpoint := ezstack.DeviceAndEndpoint{
		NetworkAddress: dev.ZigbeeDevice.NetworkAddress,
		EndpointId:     ezstack.DefaultSingleEndpointId, // incoming commands only modify the default endpoint
	}

	// we must echo the changes made back to the MQTT network (Home Assistant expects that in
	// "non-optimistic" mode)
	if err := updateAttributesAndNotifyMQTT(dev, mqttPublish, mqttPrefix, endpoint.EndpointId, func(actx *hubtypes.AttrsCtx) error {
		// when tweaking color temp, incoming MQTT message will happily ask us to:
		//
		//   {"state":"ON","color_temp":313}
		//
		// even if the state is already on. we can't therefore take state=ON at face value
		// to send "turn on" to Zigbee network (if we want to avoid unnecessary traffic). so
		// we'll do desired state vs. actual state diffs to determine which messages we'll
		// actually send.
		desiredState := &hubtypes.AttrsCtx{ // TODO: new somewhere else
			Attrs:       hubtypes.NewAttributes(),
			AttrBuilder: actx.AttrBuilder,
		}

		currentlyOn := func() bool { // toggle msg needs this
			if actx.Attrs.On != nil {
				return actx.Attrs.On.Value
			} else {
				return false // assume off if we don't know (or entity is not "on-able")
			}
		}()

		// *desiredState* will now contain {"On":true,"ColorTemperature":313}
		if err := homeassistantmqtt.MessageToAttributes(inboundMsg, desiredState, currentlyOn); err != nil {
			return err
		}

		// assigns desired state attrs to current state only if the values are different
		// (along with last change timestamp as *now*, so changed() helper can detect it)
		desiredState.Attrs.CopyDifferentAttrsTo(actx.Attrs)

		now := actx.Reported
		attrs := actx.Attrs // shorthand

		changed := func(attr hubtypes.Attribute) bool { // helper
			if isNilInterface(attr) {
				return false
			}

			return attr.LastChange().Equal(now)
		}

		if changed(attrs.On) {
			if attrs.On.Value {
				if err := zigbee.LocalCommand(endpoint, &cluster.GenOnOffOnCommand{}); err != nil {
					return err
				}
			} else {
				if err := zigbee.LocalCommand(endpoint, &cluster.GenOnOffOffCommand{}); err != nil {
					return err
				}
			}
		}

		if changed(attrs.Brightness) {
			if err := zigbee.LocalCommand(endpoint, &cluster.MoveToLevelCommand{
				Level:          uint8(attrs.Brightness.Value),
				TransitionTime: cluster.TransitionTimeFrom(1 * time.Second),
			}); err != nil {
				return err
			}
		}

		if changed(attrs.Color) {
			// 3rd return is luminance, which a color technically doesn't have
			X, Y, _ := attrs.Color.Converter().Xyz()

			if err := zigbee.LocalCommand(endpoint, &cluster.LightingColorCtrlMoveToColor{
				X:              uint16(X * 65279),
				Y:              uint16(Y * 65279),
				TransitionTime: cluster.TransitionTimeFrom(1 * time.Second),
			}); err != nil {
				return err
			}
		}

		if changed(attrs.ColorTemperature) {
			if err := zigbee.LocalCommand(endpoint, &cluster.LightingColorCtrlMoveToColorTemperature{
				uint16(attrs.ColorTemperature.Value),
				cluster.TransitionTimeFrom(1 * time.Second),
			}); err != nil {
				return err
			}
		}

		if changed(attrs.ShadePosition) {
			if err := zigbee.LocalCommand(endpoint, &cluster.ClosuresWindowCoveringGoToLiftPercentage{
				uint8(attrs.ShadePosition.Value),
			}); err != nil {
				return err
			}
		}

		if changed(attrs.ShadeStop) {
			if err := zigbee.LocalCommand(endpoint, &cluster.ClosuresWindowCoveringStop{}); err != nil {
				return err
			}
		}

		if changed(attrs.AlertSelect) {
			if err := zigbee.LocalCommand(endpoint, &cluster.GenIdentifyTriggerEffectCommand{
				Effect: cluster.EffectIdBlink,
			}); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// wrapper for when you want to update device's attributes and send any possible changes in zigbee2mqtt format
func updateAttributesAndNotifyMQTT(
	wdev *hubtypes.Device,
	mqttPublish chan<- homeassistantmqtt.Message,
	mqttPrefix string,
	endpoint zigbee.EndpointId,
	updater func(actx *hubtypes.AttrsCtx) error,
) error {
	now := time.Now().UTC()

	attrs, found := wdev.State.EndpointAttrs[endpoint]
	if !found {
		return fmt.Errorf("received message from non-declared endpoint: %d", endpoint)
	}

	actx := &hubtypes.AttrsCtx{hubtypes.NewAttrBuilder(now), attrs, endpoint}

	// this is expected to modify device's attributes
	if err := updater(actx); err != nil {
		return err
	}

	// TODO: only update attributes to node database ("master data") here, because only
	//       here we know that all Zigbee network comms were successfully ACKed

	// TODO: should this be updated?
	// wdev.LastHeard = now

	// attributes have changed - we'll need to publish updates over MQTT
	changedAttributesMsg, err := homeassistantmqtt.MessageFromChangedAttributes(
		attrs,
		wdev,
		deviceadapters.For(wdev.ZigbeeDevice).BatteryType(),
		now)
	if err != nil {
		return err
	}

	msg := homeassistantmqtt.Message{
		Topic:   mqttPrefix + "/" + wdev.ZigbeeDevice.IEEEAddress.HexPrefixedString(),
		Content: changedAttributesMsg,
	}

	select {
	case mqttPublish <- msg:
		return nil
	default:
		return errors.New("mqttPublish full")
	}
}

// https://www.home-assistant.io/docs/mqtt/discovery/
func homeAssistantAutoDiscovery(
	mqttAddr string,
	mqttPrefix string,
	nodeDatabase *nodeDb,
	logger *log.Logger,
) error {
	entities := []*homeassistant.Entity{}

	if err := nodeDatabase.withLock(func() error {
		for _, dev := range nodeDatabase.Devices {
			entities = append(entities, homeassistantmqtt.AutodiscoveryEntities(dev, mqttPrefix)...)
		}

		return nil
	}); err != nil {
		return err
	}

	if len(entities) == 0 { // nothing to do
		return nil
	}

	clientId := mqttPrefix // use prefix as client ID

	homeAssistant, mqttTask := homeassistant.NewMQTTClient(
		homeassistant.MQTTConfig{
			Address: mqttAddr,
		},
		clientId,
		logex.Levels(logger))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tasks := taskrunner.New(ctx, logger)
	tasks.Start("mqtt", mqttTask)

	if err := homeAssistant.AutodiscoverEntities(entities...); err != nil {
		return err
	}

	cancel()

	return tasks.Wait()
}

func isNilInterface(i interface{}) bool {
	return i == nil || reflect.ValueOf(i).IsNil()
}
