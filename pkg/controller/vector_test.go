package controller

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/locations"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleVector_CallsignNotOnScope(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)

	h.ctrl.HandleVector(h.ctx, &brevity.VectorRequest{
		Callsign: "eagle 1",
		Location: "home plate",
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.VectorResponse)
	require.True(t, ok)
	assert.Equal(t, "home plate", resp.Location)
	assert.False(t, resp.Contact)
	assert.False(t, resp.Status)
	assert.Nil(t, resp.Vector)
	assert.Nil(t, resp.BRA)
}

func TestHandleVector_LocationNotConfigured(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.1, 40.1})

	h.ctrl.HandleVector(h.ctx, &brevity.VectorRequest{
		Callsign: "eagle 1",
		Location: "atlantis",
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.VectorResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	assert.Equal(t, "atlantis", resp.Location)
	assert.True(t, resp.Contact)
	assert.False(t, resp.Status)
	assert.Nil(t, resp.Vector)
	assert.Nil(t, resp.BRA)
}

func TestHandleVector_HappyPath(t *testing.T) {
	t.Parallel()
	locs := []locations.Location{
		{Names: []string{"home plate"}, Longitude: 30.0, Latitude: 40.0},
	}
	h := newControllerTestHarness(t, locs)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.1, 40.1})

	h.ctrl.HandleVector(h.ctx, &brevity.VectorRequest{
		Callsign: "eagle 1",
		Location: "home plate",
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.VectorResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	assert.Equal(t, "home plate", resp.Location)
	assert.True(t, resp.Contact)
	assert.True(t, resp.Status)
	require.NotNil(t, resp.Vector)
	assert.InDelta(t, 211.0, resp.Vector.Bearing().Degrees(), bearingDeltaDegrees)
	assert.InDelta(t, 8.0, resp.Vector.Range().NauticalMiles(), rangeDeltaNauticalMiles)
	assert.Nil(t, resp.BRA)
}

func TestHandleVector_Tanker_NoCompatibleTanker(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.1, 40.1})

	h.ctrl.HandleVector(h.ctx, &brevity.VectorRequest{
		Callsign: "eagle 1",
		Location: brevity.LocationTanker,
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.VectorResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	assert.Equal(t, brevity.LocationTanker, resp.Location)
	assert.True(t, resp.Contact)
	assert.False(t, resp.Status)
	assert.Nil(t, resp.Vector)
	assert.Nil(t, resp.BRA)
}

func TestHandleVector_Tanker_FlyingBoomReceiverMatchesBoomTanker(t *testing.T) {
	t.Parallel()
	// A-10C requires flying-boom refueling. Make a boom-compatible KC-135
	// nearby and a probe-drogue KC135MPRS farther away: the A-10 should get
	// vectored to the KC-135 even though it's further from the MPRS.
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Warthog 1 Reaper", acmiA10C, coalitions.Blue, orb.Point{30.0, 40.0})
	h.insertAircraft(t, "Texaco 1", acmiKC135, coalitions.Blue, orb.Point{30.2, 40.2})
	h.insertAircraft(t, "Arco 1", acmiKC135MPRS, coalitions.Blue, orb.Point{30.1, 40.1})

	h.ctrl.HandleVector(h.ctx, &brevity.VectorRequest{
		Callsign: "warthog 1",
		Location: brevity.LocationTanker,
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.VectorResponse)
	require.True(t, ok)
	assert.Equal(t, "warthog 1", resp.Callsign)
	assert.True(t, resp.Contact)
	assert.True(t, resp.Status)
	assert.Equal(t, "Texaco 1", resp.Location)
	require.NotNil(t, resp.BRA)
	assert.InDelta(t, 31.0, resp.BRA.Bearing().Degrees(), bearingDeltaDegrees)
	assert.InDelta(t, 16.0, resp.BRA.Range().NauticalMiles(), rangeDeltaNauticalMiles)
	assert.InDelta(t, 20000.0, resp.BRA.Altitude().Feet(), altitudeDeltaFeet)
	assert.Nil(t, resp.Vector)
}

func TestHandleVector_Tanker_ProbeReceiverMatchesBasketTanker(t *testing.T) {
	t.Parallel()
	// F/A-18 is probe-and-drogue. A KC-135 (boom) should be skipped in favor
	// of the KC135MPRS (basket).
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Hornet 1 Reaper", acmiFA18C, coalitions.Blue, orb.Point{30.0, 40.0})
	h.insertAircraft(t, "Texaco 1", acmiKC135, coalitions.Blue, orb.Point{30.1, 40.1})
	h.insertAircraft(t, "Arco 1", acmiKC135MPRS, coalitions.Blue, orb.Point{30.2, 40.2})

	h.ctrl.HandleVector(h.ctx, &brevity.VectorRequest{
		Callsign: "hornet 1",
		Location: brevity.LocationTanker,
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.VectorResponse)
	require.True(t, ok)
	assert.Equal(t, "hornet 1", resp.Callsign)
	assert.True(t, resp.Contact)
	assert.True(t, resp.Status)
	assert.Equal(t, "Arco 1", resp.Location)
	require.NotNil(t, resp.BRA)
	assert.InDelta(t, 31.0, resp.BRA.Bearing().Degrees(), bearingDeltaDegrees)
	assert.InDelta(t, 16.0, resp.BRA.Range().NauticalMiles(), rangeDeltaNauticalMiles)
	assert.InDelta(t, 20000.0, resp.BRA.Altitude().Feet(), altitudeDeltaFeet)
	assert.Nil(t, resp.Vector)
}
