package controller

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/locations"
	"github.com/dharmab/skyeye/pkg/radar"
	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// vectorTestHarness sets up a Radar, Controller, and output channel for
// HandleVector tests. The Radar runs in a background goroutine and is
// stopped via t.Cleanup.
type vectorTestHarness struct {
	ctx     context.Context
	rdr     *radar.Radar
	calls   chan Call
	updates chan sim.Updated
	ctrl    *Controller
}

func newVectorTestHarness(t *testing.T, locs []locations.Location) *vectorTestHarness {
	t.Helper()
	starts := make(chan sim.Started)
	updates := make(chan sim.Updated, 16)
	fades := make(chan sim.Faded)
	rdr := radar.New(coalitions.Blue, starts, updates, fades, 25*unit.NauticalMile, 5*unit.Degree, 1*unit.NauticalMile, false, nil)
	rdr.SetBullseye(orb.Point{30.0, 40.0}, coalitions.Blue)
	rdr.SetMissionTime(time.Now())
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	var wg sync.WaitGroup
	go rdr.Run(ctx, &wg)

	ctrl := New(
		rdr,
		nil, // srsClient not exercised
		coalitions.Blue,
		false, 0,
		false, 0,
		false,
		locs,
	)
	calls := make(chan Call, 8)
	ctrl.calls = calls

	return &vectorTestHarness{
		ctx:     ctx,
		rdr:     rdr,
		calls:   calls,
		updates: updates,
		ctrl:    ctrl,
	}
}

// insertAircraft pushes updates into the radar's channel and waits for the
// trackfile to appear with a populated frame. Two updates are sent because
// the radar's handleUpdate path only applies a frame when the trackfile
// already exists.
func (h *vectorTestHarness) insertAircraft(t *testing.T, id uint64, name, acmiName string, coalition coalitions.Coalition, point orb.Point) {
	t.Helper()
	agl := 20000 * unit.Foot
	labels := trackfiles.Labels{
		ID:        id,
		Name:      name,
		Coalition: coalition,
		ACMIName:  acmiName,
	}
	frame := trackfiles.Frame{
		Time:     time.Now(),
		Point:    point,
		Altitude: 20000 * unit.Foot,
		AGL:      &agl,
		Heading:  90 * unit.Degree,
	}
	h.updates <- sim.Updated{Labels: labels, Frame: frame}
	frame.Time = frame.Time.Add(time.Second)
	h.updates <- sim.Updated{Labels: labels, Frame: frame}
	assert.Eventually(t, func() bool {
		tf := h.rdr.FindUnit(id)
		return tf != nil && !tf.IsLastKnownPointZero()
	}, time.Second, 5*time.Millisecond, "radar did not ingest trackfile for %s in time", name)
}

// expectResponse drains one response from the calls channel.
func (h *vectorTestHarness) expectResponse(t *testing.T) any {
	t.Helper()
	select {
	case c := <-h.calls:
		return c.Call
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for response")
		return nil
	}
}

func TestHandleVector_CallsignNotOnScope(t *testing.T) {
	t.Parallel()
	h := newVectorTestHarness(t, nil)

	h.ctrl.HandleVector(h.ctx, &brevity.VectorRequest{
		Callsign: "eagle 1",
		Location: "home plate",
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.VectorResponse)
	require.True(t, ok, "got %T", got)
	assert.False(t, resp.Contact)
}

func TestHandleVector_LocationNotConfigured(t *testing.T) {
	t.Parallel()
	h := newVectorTestHarness(t, nil)
	h.insertAircraft(t, 1, "Eagle 1 Reaper", "F-15C", coalitions.Blue, orb.Point{30.1, 40.1})

	h.ctrl.HandleVector(h.ctx, &brevity.VectorRequest{
		Callsign: "eagle 1",
		Location: "atlantis",
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.VectorResponse)
	require.True(t, ok, "got %T", got)
	assert.True(t, resp.Contact)
	assert.False(t, resp.Status)
}

func TestHandleVector_HappyPath(t *testing.T) {
	t.Parallel()
	locs := []locations.Location{
		{Names: []string{"home plate"}, Longitude: 30.0, Latitude: 40.0},
	}
	h := newVectorTestHarness(t, locs)
	h.insertAircraft(t, 1, "Eagle 1 Reaper", "F-15C", coalitions.Blue, orb.Point{30.1, 40.1})

	h.ctrl.HandleVector(h.ctx, &brevity.VectorRequest{
		Callsign: "eagle 1",
		Location: "home plate",
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.VectorResponse)
	require.True(t, ok, "got %T", got)
	assert.True(t, resp.Contact)
	assert.True(t, resp.Status)
	require.NotNil(t, resp.Vector)
	// Range from (30.1, 40.1) back to (30.0, 40.0) is around 7-8 NM.
	assert.InDelta(t, 8.0, resp.Vector.Range().NauticalMiles(), 4.0)
}

func TestHandleVector_Tanker_NoCompatibleTanker(t *testing.T) {
	t.Parallel()
	h := newVectorTestHarness(t, nil)
	h.insertAircraft(t, 1, "Eagle 1 Reaper", "F-15C", coalitions.Blue, orb.Point{30.1, 40.1})

	h.ctrl.HandleVector(h.ctx, &brevity.VectorRequest{
		Callsign: "eagle 1",
		Location: brevity.LocationTanker,
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.VectorResponse)
	require.True(t, ok, "got %T", got)
	assert.True(t, resp.Contact)
	assert.False(t, resp.Status)
}

func TestHandleVector_Tanker_FlyingBoomReceiverMatchesBoomTanker(t *testing.T) {
	t.Parallel()
	// A-10C requires flying-boom refueling. Make a boom-compatible KC-135
	// nearby and a probe-drogue KC135MPRS farther away: the A-10 should get
	// vectored to the KC-135 even though it's further from the MPRS.
	h := newVectorTestHarness(t, nil)
	h.insertAircraft(t, 1, "Warthog 1 Reaper", "A-10C", coalitions.Blue, orb.Point{30.0, 40.0})
	h.insertAircraft(t, 100, "Texaco 1", "KC-135", coalitions.Blue, orb.Point{30.2, 40.2})
	h.insertAircraft(t, 101, "Arco 1", "KC135MPRS", coalitions.Blue, orb.Point{30.1, 40.1})

	h.ctrl.HandleVector(h.ctx, &brevity.VectorRequest{
		Callsign: "warthog 1",
		Location: brevity.LocationTanker,
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.VectorResponse)
	require.True(t, ok, "got %T", got)
	assert.True(t, resp.Contact)
	assert.True(t, resp.Status)
	assert.Equal(t, "Texaco 1", resp.Location)
	require.NotNil(t, resp.BRA)
}

func TestHandleVector_Tanker_ProbeReceiverMatchesBasketTanker(t *testing.T) {
	t.Parallel()
	// F/A-18 is probe-and-drogue. A KC-135 (boom) should be skipped in favor
	// of the KC135MPRS (basket).
	h := newVectorTestHarness(t, nil)
	h.insertAircraft(t, 1, "Hornet 1 Reaper", "FA-18C_hornet", coalitions.Blue, orb.Point{30.0, 40.0})
	h.insertAircraft(t, 100, "Texaco 1", "KC-135", coalitions.Blue, orb.Point{30.1, 40.1})
	h.insertAircraft(t, 101, "Arco 1", "KC135MPRS", coalitions.Blue, orb.Point{30.2, 40.2})

	h.ctrl.HandleVector(h.ctx, &brevity.VectorRequest{
		Callsign: "hornet 1",
		Location: brevity.LocationTanker,
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.VectorResponse)
	require.True(t, ok, "got %T", got)
	assert.True(t, resp.Contact)
	assert.True(t, resp.Status)
	assert.Equal(t, "Arco 1", resp.Location)
}
