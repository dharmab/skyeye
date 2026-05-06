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
	_, ok := got.(brevity.NegativeRadarContactResponse)
	require.True(t, ok)
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
	assert.False(t, resp.Status)
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
	assert.True(t, resp.Status)
	require.NotNil(t, resp.Group)
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
	assert.False(t, resp.Status)
}
