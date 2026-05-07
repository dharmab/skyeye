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

func TestHandleStrobe_CallsignNotOnRadar(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)

	h.ctrl.HandleStrobe(h.ctx, &brevity.StrobeRequest{
		Callsign: "eagle 1",
		Bearing:  bearings.NewMagneticBearing(90 * unit.Degree),
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.NegativeRadarContactResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
}

func TestHandleStrobe_NoHostileInCone(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})

	h.ctrl.HandleStrobe(h.ctx, &brevity.StrobeRequest{
		Callsign: "eagle 1",
		Bearing:  bearings.NewMagneticBearing(90 * unit.Degree),
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.StrobeResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	assert.InDelta(t, 90.0, resp.Bearing.Degrees(), 0.1)
	assert.False(t, resp.Status)
	assert.Nil(t, resp.Group)
}

func TestHandleStrobe_HostileInCone(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})
	h.insertAircraft(t, "Bandit 1", acmiSu27, coalitions.Red, orb.Point{30.5, 40.0})

	h.ctrl.HandleStrobe(h.ctx, &brevity.StrobeRequest{
		Callsign: "eagle 1",
		Bearing:  bearings.NewMagneticBearing(90 * unit.Degree),
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.StrobeResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	assert.InDelta(t, 90.0, resp.Bearing.Degrees(), 0.1)
	assert.True(t, resp.Status)
	require.NotNil(t, resp.Group)
	assert.Equal(t, brevity.Hostile, resp.Group.Declaration())
	assert.Equal(t, 1, resp.Group.Contacts())
	require.NotNil(t, resp.Group.BRAA())
	assert.InDelta(t, 84.0, resp.Group.BRAA().Bearing().Degrees(), bearingDeltaDegrees)
	assert.InDelta(t, 23.0, resp.Group.BRAA().Range().NauticalMiles(), rangeDeltaNauticalMiles)
	assert.InDelta(t, 20000.0, resp.Group.BRAA().Altitude().Feet(), altitudeDeltaFeet)
	assert.Contains(t, resp.Group.Platforms(), "Flanker")
}

func TestHandleStrobe_HostileOutsideCone(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})
	// Place hostile roughly north — well outside 30-degree cone of magnetic bearing 090
	h.insertAircraft(t, "Bandit 1", acmiSu27, coalitions.Red, orb.Point{30.0, 40.5})

	h.ctrl.HandleStrobe(h.ctx, &brevity.StrobeRequest{
		Callsign: "eagle 1",
		Bearing:  bearings.NewMagneticBearing(90 * unit.Degree),
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.StrobeResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	assert.InDelta(t, 90.0, resp.Bearing.Degrees(), 0.1)
	assert.False(t, resp.Status)
	assert.Nil(t, resp.Group)
}
