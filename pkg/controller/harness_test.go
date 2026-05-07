package controller

import (
	"context"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/locations"
	"github.com/dharmab/skyeye/pkg/radar"
	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
)

// controllerTestHarness sets up a Radar, Controller, and output channel for
// handler tests. The Radar runs in a background goroutine and is stopped via
// t.Cleanup.
type controllerTestHarness struct {
	ctx     context.Context
	rdr     *radar.Radar
	calls   chan Call
	updates chan sim.Updated
	ctrl    *Controller
	nextID  uint64
}

func newControllerTestHarness(t *testing.T, locs []locations.Location) *controllerTestHarness {
	t.Helper()
	starts := make(chan sim.Started)
	updates := make(chan sim.Updated, 16)
	fades := make(chan sim.Faded)
	rdr := radar.New(coalitions.Blue, starts, updates, fades, 25*unit.NauticalMile, 5*unit.Degree, 1*unit.NauticalMile, false, nil)
	rdr.SetBullseye(orb.Point{30.0, 40.0}, coalitions.Blue)
	rdr.SetBullseye(orb.Point{35.0, 33.0}, coalitions.Red)
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

	return &controllerTestHarness{
		ctx:     ctx,
		rdr:     rdr,
		calls:   calls,
		updates: updates,
		ctrl:    ctrl,
	}
}

const (
	bearingDeltaDegrees     = 5.0
	rangeDeltaNauticalMiles = 1.0
	altitudeDeltaFeet       = 1000.0
)

const (
	acmiF15C      = "F-15C"
	acmiF16C      = "F-16C_50"
	acmiFA18C     = "FA-18C_hornet"
	acmiA10C      = "A-10C"
	acmiF14B      = "F-14B"
	acmiSu27      = "Su-27"
	acmiSu25T     = "Su-25T"
	acmiMiG29A    = "MiG-29A"
	acmiMiG21     = "MiG-21Bis"
	acmiJ11A      = "J-11A"
	acmiKa50      = "Ka-50"
	acmiKC135     = "KC-135"
	acmiKC135MPRS = "KC135MPRS"
)

type insertOption func(*insertConfig)

type insertConfig struct {
	heading unit.Angle
}

func withHeading(heading unit.Angle) insertOption {
	return func(c *insertConfig) {
		c.heading = heading
	}
}

// insertAircraft pushes updates into the radar's channel and waits for the
// trackfile to appear with a populated frame. Two updates are sent because
// the radar's handleUpdate path only applies a frame when the trackfile
// already exists.
func (h *controllerTestHarness) insertAircraft(t *testing.T, name, acmiName string, coalition coalitions.Coalition, point orb.Point, opts ...insertOption) {
	t.Helper()
	h.nextID++
	id := h.nextID
	cfg := insertConfig{heading: 90 * unit.Degree}
	for _, opt := range opts {
		opt(&cfg)
	}
	agl := 20000 * unit.Foot
	labels := trackfiles.Labels{
		ID:        id,
		Name:      name,
		Coalition: coalition,
		ACMIName:  acmiName,
	}
	// Offset the first frame slightly behind the heading so the trackfile
	// computes a non-zero ground speed and meaningful direction/aspect.
	rad := cfg.heading.Radians()
	const offset = 0.005
	prevPoint := orb.Point{
		point[0] - offset*math.Sin(rad),
		point[1] - offset*math.Cos(rad),
	}
	frame := trackfiles.Frame{
		Time:     time.Now(),
		Point:    prevPoint,
		Altitude: 20000 * unit.Foot,
		AGL:      &agl,
		Heading:  cfg.heading,
	}
	h.updates <- sim.Updated{Labels: labels, Frame: frame}
	frame.Time = frame.Time.Add(time.Second)
	frame.Point = point
	h.updates <- sim.Updated{Labels: labels, Frame: frame}
	assert.Eventually(t, func() bool {
		tf := h.rdr.FindUnit(id)
		return tf != nil && !tf.IsLastKnownPointZero()
	}, time.Second, 5*time.Millisecond, "radar did not ingest trackfile for %s in time", name)
}

// expectResponse drains one response from the calls channel.
func (h *controllerTestHarness) expectResponse(t *testing.T) any {
	t.Helper()
	select {
	case c := <-h.calls:
		return c.Call
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for response")
		return nil
	}
}
