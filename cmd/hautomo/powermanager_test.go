package main

import (
	"github.com/function61/gokit/assert"
	"github.com/function61/hautomo/pkg/hapitypes"
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

func TestPowerManagerWithExplicit(t *testing.T) {
	pm := NewPowerManager()
	pm.Register("dev", true)

	pm.Set("dev", hapitypes.PowerKindOn) // should not do anything
	assert.Assert(t, len(pm.Diff()) == 0)

	pm.SetExplicit("dev", hapitypes.PowerKindOn)
	pd := pm.Diff()
	assert.EqualString(t, serialize(pd), "dev => on")
	pm.ApplyDiff(pd[0])
	assert.Assert(t, len(pm.Diff()) == 0)

	pm.Set("dev", hapitypes.PowerKindOn)
	assert.Assert(t, len(pm.Diff()) == 0)

	pm.SetExplicit("dev", hapitypes.PowerKindOn)
	assert.EqualString(t, serialize(pm.Diff()), "dev => on")
}

func TestPowerManagerSetBypassingDiffs(t *testing.T) {
	pm := NewPowerManager()
	pm.Register("dev", true)

	assert.Assert(t, len(pm.Diff()) == 0)

	// Set(off) would normally cause diff
	pm.SetBypassingDiffs("dev", hapitypes.PowerKindOff)

	assert.Assert(t, len(pm.Diff()) == 0)
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
