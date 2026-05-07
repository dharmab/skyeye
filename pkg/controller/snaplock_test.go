package controller

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleSnaplock_CallsignNotOnRadar(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)

	h.ctrl.HandleSnaplock(h.ctx, &brevity.SnaplockRequest{
		Callsign: "eagle 1",
		BRA:      brevity.NewBRA(bearings.NewMagneticBearing(90*unit.Degree), 20*unit.NauticalMile, 20000*unit.Foot),
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.NegativeRadarContactResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
}

func TestHandleSnaplock_Clean(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})

	// Point BRA at empty space far to the east
	h.ctrl.HandleSnaplock(h.ctx, &brevity.SnaplockRequest{
		Callsign: "eagle 1",
		BRA:      brevity.NewBRA(bearings.NewMagneticBearing(90*unit.Degree), 80*unit.NauticalMile, 20000*unit.Foot),
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.SnaplockResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	assert.Equal(t, brevity.Clean, resp.Declaration)
	assert.Nil(t, resp.Group)
}

func TestHandleSnaplock_Friendly(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})
	// Place friendly ~25nm east
	h.insertAircraft(t, "Eagle 2 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.5, 40.0})

	h.ctrl.HandleSnaplock(h.ctx, &brevity.SnaplockRequest{
		Callsign: "eagle 1",
		BRA:      brevity.NewBRA(bearings.NewMagneticBearing(90*unit.Degree), 25*unit.NauticalMile, 20000*unit.Foot),
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.SnaplockResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	assert.Equal(t, brevity.Friendly, resp.Declaration)
	require.NotNil(t, resp.Group)
	assert.Equal(t, brevity.Friendly, resp.Group.Declaration())
}

func TestHandleSnaplock_Hostile(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})
	// Place hostile ~25nm east
	h.insertAircraft(t, "Bandit 1", acmiSu27, coalitions.Red, orb.Point{30.5, 40.0})

	h.ctrl.HandleSnaplock(h.ctx, &brevity.SnaplockRequest{
		Callsign: "eagle 1",
		BRA:      brevity.NewBRA(bearings.NewMagneticBearing(90*unit.Degree), 25*unit.NauticalMile, 20000*unit.Foot),
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.SnaplockResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	assert.Equal(t, brevity.Hostile, resp.Declaration)
	require.NotNil(t, resp.Group)
	assert.Equal(t, brevity.Hostile, resp.Group.Declaration())
	assert.Equal(t, 1, resp.Group.Contacts())
	require.NotNil(t, resp.Group.BRAA())
	assert.InDelta(t, 84.0, resp.Group.BRAA().Bearing().Degrees(), bearingDeltaDegrees)
	assert.InDelta(t, 23.0, resp.Group.BRAA().Range().NauticalMiles(), rangeDeltaNauticalMiles)
	assert.InDelta(t, 20000.0, resp.Group.BRAA().Altitude().Feet(), altitudeDeltaFeet)
	assert.Equal(t, brevity.Aspect(brevity.Drag), resp.Group.BRAA().Aspect())
	assert.Contains(t, resp.Group.Platforms(), "Flanker")
}

func TestHandleSnaplock_Furball(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})
	// Place friendly and hostile near same location ~25nm east
	h.insertAircraft(t, "Eagle 2 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.5, 40.0})
	h.insertAircraft(t, "Bandit 1", acmiSu27, coalitions.Red, orb.Point{30.5, 40.01})

	h.ctrl.HandleSnaplock(h.ctx, &brevity.SnaplockRequest{
		Callsign: "eagle 1",
		BRA:      brevity.NewBRA(bearings.NewMagneticBearing(90*unit.Degree), 25*unit.NauticalMile, 20000*unit.Foot),
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.SnaplockResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	assert.Equal(t, brevity.Furball, resp.Declaration)
	assert.Nil(t, resp.Group)
}

func TestHandleSnaplock_HostilePrefersHotAspect(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	// Fighter at bullseye
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})
	// Hostile heading east (away from fighter) → Drag aspect from BRA 090
	h.insertAircraft(t, "Bandit 1", acmiSu27, coalitions.Red, orb.Point{30.5, 40.0})
	// Hostile heading west (toward fighter) → Hot aspect from BRA 090
	// Placed >3nm from Bandit 1 so they form separate groups, but within 10nm of the BRA point
	h.insertAircraft(t, "Bandit 2", acmiMiG29A, coalitions.Red, orb.Point{30.6, 40.08}, withHeading(270*unit.Degree))

	h.ctrl.HandleSnaplock(h.ctx, &brevity.SnaplockRequest{
		Callsign: "eagle 1",
		BRA:      brevity.NewBRA(bearings.NewMagneticBearing(90*unit.Degree), 25*unit.NauticalMile, 20000*unit.Foot),
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.SnaplockResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	assert.Equal(t, brevity.Hostile, resp.Declaration)
	require.NotNil(t, resp.Group)
	assert.Equal(t, brevity.Hostile, resp.Group.Declaration())
	assert.Equal(t, 1, resp.Group.Contacts())
	require.NotNil(t, resp.Group.BRAA())
	assert.Equal(t, brevity.Aspect(brevity.Hot), resp.Group.BRAA().Aspect())
	assert.Equal(t, brevity.Aspect(brevity.Hot), resp.Group.Aspect())
	assert.Contains(t, resp.Group.Platforms(), "Fulcrum")
}
