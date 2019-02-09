package main

import (
	"github.com/function61/hautomo/pkg/hapitypes"
	"time"
)

// policy idea: https://twitter.com/bradfitz/status/1056736707477819392
type policyEngine struct {
	booleans            *booleanStorage
	kitchenMotionSensor *hapitypes.Device
}

func newPolicyEngine(booleans *booleanStorage, kitchenMotionSensor *hapitypes.Device) *policyEngine {
	return &policyEngine{
		booleans:            booleans,
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

	powerManager.Set("kitchenLight", boolToPowerKind(p.shouldKitchenLightBeOn()))
}

func (p *policyEngine) shouldKitchenLightBeOn() bool {
	if anybodyHome, _ := p.booleans.Get("anybodyHome"); !anybodyHome {
		return false
	}

	if environmentHasLight, _ := p.booleans.Get("environmentHasLight"); environmentHasLight {
		return false
	}

	lastMotion := p.kitchenMotionSensor.LastMotion
	if lastMotion == nil {
		return false
	}

	twoMinutesAgo := time.Now().Add(-2 * time.Minute)

	return lastMotion.After(twoMinutesAgo)
}
