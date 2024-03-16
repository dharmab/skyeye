package radar

import (
	"math"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/trackfile"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/planar"

	"github.com/paulmach/orb/geo"
)

type group struct {
	isThreat bool
	contacts []*trackfile.Trackfile
	bullseye orb.Point
	platform string
}

var _ brevity.Group = group{}

func newGroupUsingBullseye(bullseye orb.Point) group {
	return group{
		bullseye: bullseye,
		contacts: make([]*trackfile.Trackfile, 0),
	}
}

func (g group) Threat() bool {
	return g.isThreat
}

func (g group) Contacts() int {
	return len(g.contacts)
}

func (g group) Bullseye() *brevity.Bullseye {
	mp := orb.MultiPoint{}
	for _, tf := range g.contacts {
		mp = append(mp, tf.Track.Front().Point)
	}
	center := mp.Bound().Center()

	bearing := unit.Angle(geo.Bearing(g.bullseye, center)) * unit.Degree
	distance := unit.Length(planar.Distance(g.bullseye, center)) * unit.Meter
	return brevity.NewBullseye(bearing, distance)
}

func (g group) Altitude() unit.Length {
	var sum unit.Length
	for _, tf := range g.contacts {
		sum += tf.Track.Front().Altitude
	}
	mean := sum / unit.Length(len(g.contacts))
	rounded := unit.Length((math.Round(mean.Feet()/1000) * 1000)) * unit.Foot
	return rounded
}

func (g group) Track() brevity.Track {
	return brevity.UnknownDirection
}

func (g group) Aspect() brevity.Aspect {
	return brevity.UnknownAspect
}

func (g group) BRAA() brevity.BRAA {
	return nil
}

func (g group) Declaration() brevity.Declaration {
	return brevity.Unable
}

func (g group) Heavy() bool {
	return len(g.contacts) >= 3
}

func (g group) Platform() string {
	return g.platform
}

func (g group) High() bool {
	return g.Altitude() > 40000*unit.Foot
}

func (g group) Fast() bool {
	return false
}

func (g group) VeryFast() bool {
	return false
}
