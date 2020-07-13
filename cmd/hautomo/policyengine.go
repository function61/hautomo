package main

import (
	"time"

	"github.com/function61/hautomo/pkg/hapitypes"
)

var (
	truep  = func(inp bool) *bool { return &inp }(true)
	falsep = func(inp bool) *bool { return &inp }(false)
)

// policy idea: https://twitter.com/bradfitz/status/1056736707477819392
type policyEngine struct {
	booleans *booleanStorage

	// control group
	kitchenLight        *hapitypes.Device
	kitchenMotionSensor *hapitypes.Device

	// control group
	bathroomCabinetLight *hapitypes.Device
	bathroomMotionSensor *hapitypes.Device
	bathroomDoor         *hapitypes.Device

	// control group
	mirrorLight         *hapitypes.Device
	bedroomMotionSensor *hapitypes.Device
}

// obtain won't be called after this ctor returns
func newPolicyEngine(booleans *booleanStorage, obtain func(key string) *hapitypes.Device) *policyEngine {
	return &policyEngine{
		booleans:             booleans,
		kitchenLight:         obtain("kitchenLight"),
		kitchenMotionSensor:  obtain("kitchenMotion"),
		bathroomCabinetLight: obtain("bathroomCabinetLight"),
		bathroomMotionSensor: obtain("bathroomMotion"),
		bathroomDoor:         obtain("bathroomDoor"),
		mirrorLight:          obtain("mirrorLight"),
		bedroomMotionSensor:  obtain("bedroomMotion"),
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

	on = p.shouldBathroomCabinetLightBeOn()
	if on != nil { // is nil if we don't want to act
		powerManager.Set("bathroomCabinetLight", boolToPowerKind(*on))
	}

	on = p.shouldMirrorLightBeOn()
	if on != nil { // is nil if we don't want to act
		powerManager.Set("mirrorLight", boolToPowerKind(*on))
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
		return falsep
	}

	if environmentHasLight, _ := p.booleans.Get("environmentHasLight"); environmentHasLight {
		return falsep
	}

	lastMotion := p.kitchenMotionSensor.LastMotion
	if lastMotion == nil {
		return falsep
	}

	twoMinutesAgo := now.Add(-2 * time.Minute)

	return bptr(lastMotion.After(twoMinutesAgo))
}

func (p *policyEngine) shouldBathroomCabinetLightBeOn() *bool {
	now := time.Now()

	if p.bathroomCabinetLight.LastExplicitPowerEvent != nil {
		fifteenMinutesAgo := now.Add(-15 * time.Minute)

		if p.bathroomCabinetLight.LastExplicitPowerEvent.After(fifteenMinutesAgo) {
			return nil
		}
	}

	if anybodyHome, _ := p.booleans.Get("anybodyHome"); !anybodyHome {
		return falsep
	}

	lastMotion := p.bathroomMotionSensor.LastMotion
	if lastMotion == nil {
		return falsep
	}

	threeMinutesAgo := now.Add(-5 * time.Minute)
	dayAgo := now.Add(-24 * time.Hour)

	// light should remain on if door was closed and movement detected after that (= it
	// means that someone must be present in the room since that contact sensor is the only egress)
	lastContact := p.bathroomDoor.LastContact
	if lastContact != nil && lastContact.Contact && lastContact.When.After(dayAgo) && lastMotion.After(lastContact.When) {
		return truep
	}

	return bptr(lastMotion.After(threeMinutesAgo))
}

func (p *policyEngine) shouldMirrorLightBeOn() *bool {
	now := time.Now()

	if p.mirrorLight.LastExplicitPowerEvent != nil {
		twelveHoursAgo := now.Add(-12 * time.Hour)

		if p.mirrorLight.LastExplicitPowerEvent.After(twelveHoursAgo) {
			return nil
		}
	}

	if environmentHasLight, _ := p.booleans.Get("environmentHasLight"); environmentHasLight {
		return falsep
	}

	if anybodyHome, _ := p.booleans.Get("anybodyHome"); !anybodyHome {
		return falsep
	}

	lastMotion := p.bedroomMotionSensor.LastMotion
	if lastMotion == nil {
		return falsep
	}

	threeMinutesAgo := now.Add(-3 * time.Minute)

	return bptr(lastMotion.After(threeMinutesAgo))
}

func bptr(b bool) *bool {
	if b {
		return truep
	} else {
		return falsep
	}
}
