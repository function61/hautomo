package main

import (
	"github.com/function61/hautomo/pkg/hapitypes"
)

type PowerDiff struct {
	Device string
	On     bool
}

type PowerManager struct {
	desired map[string]bool
	actual  map[string]bool
}

// implements desired state reconciliation for controlling device's power
func NewPowerManager() *PowerManager {
	return &PowerManager{
		desired: map[string]bool{},
		actual:  map[string]bool{},
	}
}

func (p *PowerManager) GetActual(deviceId string) bool {
	return p.actual[deviceId]
}

func (p *PowerManager) Register(deviceId string, isOn bool) {
	p.desired[deviceId] = isOn
	p.actual[deviceId] = isOn
}

func (p *PowerManager) SetExplicit(deviceId string, power hapitypes.PowerKind) {
	p.Set(deviceId, power)

	// for explicit sets, we want to always force a diff. this hack does it
	p.actual[deviceId] = !p.desired[deviceId]
}

func (p *PowerManager) Set(deviceId string, power hapitypes.PowerKind) {
	p.desired[deviceId] = p.getDesired(deviceId, power)
}

// just set actual state to something without triggering a diff that would send a message.
// one case for wanting this is sending brightness msg which will implicitly turn the bulb
// on even though we didn't send an explicit power message
func (p *PowerManager) SetBypassingDiffs(deviceId string, power hapitypes.PowerKind) {
	p.Set(deviceId, power)

	p.actual[deviceId] = p.desired[deviceId]
}

func (p *PowerManager) getDesired(deviceId string, power hapitypes.PowerKind) bool {
	switch power {
	case hapitypes.PowerKindOn:
		return true
	case hapitypes.PowerKindOff:
		return false
	case hapitypes.PowerKindToggle:
		return !p.actual[deviceId]
	default:
		panic("unknown PowerKind")
	}
}

// marks diff as "committed" so it won't show up as diff the next time
func (p *PowerManager) CommitDiff(pd PowerDiff) {
	p.actual[pd.Device] = pd.On
}

func (p *PowerManager) Diff() []PowerDiff {
	diff := []PowerDiff{}
	for deviceId, isDesiredOn := range p.desired {
		isActuallyOn := p.actual[deviceId]
		if isDesiredOn != isActuallyOn {
			diff = append(diff, PowerDiff{deviceId, isDesiredOn})
		}
	}

	return diff
}
