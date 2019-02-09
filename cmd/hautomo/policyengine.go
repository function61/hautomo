package main

import (
	"github.com/function61/hautomo/pkg/hapitypes"
	"time"
)

// policy idea: https://twitter.com/bradfitz/status/1056736707477819392
type policyEngine struct {
	booleans            *booleanStorage
	kitchenLight        *hapitypes.Device
	kitchenMotionSensor *hapitypes.Device
}

func newPolicyEngine(
	booleans *booleanStorage,
	kitchenLight *hapitypes.Device,
	kitchenMotionSensor *hapitypes.Device,
) *policyEngine {
	return &policyEngine{
		booleans:            booleans,
		kitchenLight:        kitchenLight,
		kitchenMotionSensor: kitchenMotionSensor,
	}
}

func (p *policyEngine) evaluatePowerPolicies(powerManager *PowerManager) {
	boolToPowerKind := func(on bool) hapitypes.PowerKind {
		if on {
			return hapitypes.PowerKindOn
		} else {
			return hapitypes.PowerKindOff
		}
	}

	on := p.shouldKitchenLightBeOn()
	if on != nil { // is nil if we don't want to act
		powerManager.Set("kitchenLight", boolToPowerKind(*on))
	}
}

func (p *policyEngine) shouldKitchenLightBeOn() *bool {
	now := time.Now()

	if p.kitchenLight.LastExplicitPowerEvent != nil {
		fifteenMinutesAgo := now.Add(-15 * time.Minute)

		if p.kitchenLight.LastExplicitPowerEvent.After(fifteenMinutesAgo) {
			return nil
		}
	}

	if anybodyHome, _ := p.booleans.Get("anybodyHome"); !anybodyHome {
		return bptr(false)
	}

	if environmentHasLight, _ := p.booleans.Get("environmentHasLight"); environmentHasLight {
		return bptr(false)
	}

	lastMotion := p.kitchenMotionSensor.LastMotion
	if lastMotion == nil {
		return bptr(false)
	}

	twoMinutesAgo := now.Add(-2 * time.Minute)

	return bptr(lastMotion.After(twoMinutesAgo))
}

func bptr(b bool) *bool {
	return &b
}
