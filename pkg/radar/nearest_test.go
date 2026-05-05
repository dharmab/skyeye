package radar

import (
	"testing"
	"time"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/encyclopedia"
	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestRadarWithContacts constructs a Radar suitable for unit tests that
// only exercise the contact database. The channels are never read from.
func newTestRadarWithContacts() *Radar {
	starts := make(chan sim.Started)
	updates := make(chan sim.Updated)
	fades := make(chan sim.Faded)
	return New(coalitions.Blue, starts, updates, fades, 25*unit.NauticalMile, 5*unit.Degree, 1*unit.NauticalMile, false, nil)
}

// insertTanker adds a tanker trackfile at the given point to the radar's
// contact database.
func insertTanker(t *testing.T, r *Radar, id uint64, name, acmiName string, coalition coalitions.Coalition, point orb.Point) {
	t.Helper()
	tf := trackfiles.New(trackfiles.Labels{
		ID:        id,
		Name:      name,
		Coalition: coalition,
		ACMIName:  acmiName,
	})
	agl := 25000 * unit.Foot
	tf.Update(trackfiles.Frame{
		Time:     time.Now(),
		Point:    point,
		Altitude: 25000 * unit.Foot,
		AGL:      &agl,
		Heading:  90 * unit.Degree,
	})
	r.contacts.set(tf)
}

func TestFindNearestTanker_NoTankersInScope(t *testing.T) {
	t.Parallel()
	r := newTestRadarWithContacts()
	origin := orb.Point{30.0, 40.0}
	got := r.FindNearestTanker(origin, coalitions.Blue, encyclopedia.FlyingBoom)
	assert.Nil(t, got)
}

func TestFindNearestTanker_OppositeCoalitionIgnored(t *testing.T) {
	t.Parallel()
	r := newTestRadarWithContacts()
	origin := orb.Point{30.0, 40.0}
	insertTanker(t, r, 1, "Texaco 1", "KC-135", coalitions.Red, orb.Point{30.1, 40.1})
	got := r.FindNearestTanker(origin, coalitions.Blue, encyclopedia.FlyingBoom)
	assert.Nil(t, got)
}

func TestFindNearestTanker_NearestWinsAmongCompatible(t *testing.T) {
	t.Parallel()
	r := newTestRadarWithContacts()
	origin := orb.Point{30.0, 40.0}
	insertTanker(t, r, 1, "Texaco 1", "KC-135", coalitions.Blue, orb.Point{31.0, 41.0})
	insertTanker(t, r, 2, "Texaco 2", "KC-135", coalitions.Blue, orb.Point{30.1, 40.1})
	got := r.FindNearestTanker(origin, coalitions.Blue, encyclopedia.FlyingBoom)
	require.NotNil(t, got)
	assert.Equal(t, uint64(2), got.Contact.ID)
}

func TestFindNearestTanker_IncompatibleMethodFiltered(t *testing.T) {
	t.Parallel()
	r := newTestRadarWithContacts()
	origin := orb.Point{30.0, 40.0}
	// KC-135 (boom-only) nearby, KC135MPRS (basket) further away.
	insertTanker(t, r, 1, "Texaco 1", "KC-135", coalitions.Blue, orb.Point{30.1, 40.1})
	insertTanker(t, r, 2, "Arco 1", "KC135MPRS", coalitions.Blue, orb.Point{31.0, 41.0})

	gotBoom := r.FindNearestTanker(origin, coalitions.Blue, encyclopedia.FlyingBoom)
	require.NotNil(t, gotBoom)
	assert.Equal(t, uint64(1), gotBoom.Contact.ID)

	gotBasket := r.FindNearestTanker(origin, coalitions.Blue, encyclopedia.ProbeAndDrogue)
	require.NotNil(t, gotBasket)
	assert.Equal(t, uint64(2), gotBasket.Contact.ID)
}

func TestFindNearestTanker_BeyondSearchRadiusFiltered(t *testing.T) {
	t.Parallel()
	r := newTestRadarWithContacts()
	origin := orb.Point{0.0, 0.0}
	// Place a tanker on the opposite side of the planet — well beyond the
	// 2400-mile search radius.
	insertTanker(t, r, 1, "Texaco 1", "KC-135", coalitions.Blue, orb.Point{179.0, 0.0})
	got := r.FindNearestTanker(origin, coalitions.Blue, encyclopedia.FlyingBoom)
	assert.Nil(t, got)
}

func TestFindNearestGroupWithBullseyeReturnsNilWhenNoTrackfile(t *testing.T) {
	t.Parallel()
	r := newTestRadarWithContacts()
	grp := r.FindNearestGroupWithBullseye(
		orb.Point{-115.0, 36.0},
		0*unit.Foot,
		50000*unit.Foot,
		100*unit.NauticalMile,
		coalitions.Red,
		brevity.Aircraft,
	)
	assert.Nil(t, grp)
}
