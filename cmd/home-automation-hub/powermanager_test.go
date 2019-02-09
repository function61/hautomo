package main

import (
	"github.com/function61/gokit/assert"
	"github.com/function61/home-automation-hub/pkg/hapitypes"
	"strings"
	"testing"
)

func TestPowerManager(t *testing.T) {
	pm := NewPowerManager()
	pm.Register("foo", false)
	pm.Register("bar", false)

	assert.Assert(t, pm.GetActual("foo") == false)
	assert.Assert(t, len(pm.Diff()) == 0)

	pm.Set("foo", hapitypes.PowerKindOn)
	assert.Assert(t, pm.GetActual("foo") == false)

	diff := pm.Diff()
	assert.Assert(t, len(diff) == 1)
	assert.EqualString(t, serialize(diff), "foo => on")

	pm.ApplyDiff(diff[0])
	assert.Assert(t, pm.GetActual("foo") == true)

	assert.Assert(t, len(pm.Diff()) == 0)

	pm.Set("foo", hapitypes.PowerKindOn)
	assert.Assert(t, len(pm.Diff()) == 0)

	pm.Set("foo", hapitypes.PowerKindToggle)
	assert.EqualString(t, serialize(pm.Diff()), "foo => off")
}

func TestPowerManagerWithDeviceGroup(t *testing.T) {
	pm := NewPowerManager()
	pm.Register("notGroup", true)
	pm.RegisterDeviceGroup("isGroup", true)

	pm.Set("notGroup", hapitypes.PowerKindOn) // should not do anything
	assert.Assert(t, len(pm.Diff()) == 0)

	pm.Set("isGroup", hapitypes.PowerKindOn)
	pd := pm.Diff()
	assert.EqualString(t, serialize(pd), "isGroup => on")
	pm.ApplyDiff(pd[0])
	assert.Assert(t, len(pm.Diff()) == 0)

	pm.Set("isGroup", hapitypes.PowerKindOn) // asking for on again should again yield diff
	assert.EqualString(t, serialize(pm.Diff()), "isGroup => on")
}

func serialize(diffs []PowerDiff) string {
	serialized := []string{}

	for _, diff := range diffs {
		if diff.On {
			serialized = append(serialized, diff.Device+" => on")
		} else {
			serialized = append(serialized, diff.Device+" => off")
		}
	}

	return strings.Join(serialized, ", ")
}
