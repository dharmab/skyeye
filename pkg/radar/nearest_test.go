package radar

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
)

func TestFindNearestGroupWithBullseyeReturnsNilWhenNoTrackfile(t *testing.T) {
	t.Parallel()

	r := New(coalitions.Blue, nil, nil, nil, 0, false)
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
