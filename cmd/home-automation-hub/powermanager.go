package main

import (
	"github.com/function61/home-automation-hub/pkg/hapitypes"
)

type PowerDiff struct {
	Device string
	On     bool
}

type PowerManager struct {
	desired       map[string]bool
	actual        map[string]bool
	isDeviceGroup map[string]bool
}

// implements desired state reconciliation for controlling device's power
func NewPowerManager() *PowerManager {
	return &PowerManager{
		desired:       map[string]bool{},
		actual:        map[string]bool{},
		isDeviceGroup: map[string]bool{},
	}
}

func (p *PowerManager) GetActual(deviceId string) bool {
	return p.actual[deviceId]
}

func (p *PowerManager) Register(deviceId string, isOn bool) {
	p.desired[deviceId] = isOn
	p.actual[deviceId] = isOn
}

func (p *PowerManager) RegisterDeviceGroup(deviceId string, isOn bool) {
	p.Register(deviceId, isOn)

	p.isDeviceGroup[deviceId] = true
}

func (p *PowerManager) Set(deviceId string, power hapitypes.PowerKind) {
	var desired bool
	switch power {
	case hapitypes.PowerKindOn:
		desired = true
	case hapitypes.PowerKindOff:
		desired = false
	case hapitypes.PowerKindToggle:
		desired = !p.actual[deviceId]
	default:
		panic("unknown PowerKind")
	}

	// if this was a device group, hack actual as different so diffs will get always applied
	if _, isDeviceGroup := p.isDeviceGroup[deviceId]; isDeviceGroup {
		p.actual[deviceId] = !desired
	}

	p.desired[deviceId] = desired
}

func (p *PowerManager) ApplyDiff(pd PowerDiff) {
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
