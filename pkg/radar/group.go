package radar

import (
	"math"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/trackfile"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"

	"github.com/paulmach/orb/geo"
)

type group struct {
	isThreat     bool
	contacts     []*trackfile.Trackfile
	bullseye     *orb.Point
	braa         brevity.BRAA
	platforms    []string
	aspect       *brevity.Aspect
	declaraction brevity.Declaration
}

var _ brevity.Group = &group{}

func newGroupUsingBullseye(bullseye orb.Point) *group {
	return &group{
		bullseye:     &bullseye,
		contacts:     make([]*trackfile.Trackfile, 0),
		declaraction: brevity.Unable,
	}
}

func (g *group) Threat() bool {
	return g.isThreat
}

func (g *group) SetThreat(isThreat bool) {
	g.isThreat = isThreat
}

func (g *group) Contacts() int {
	return len(g.contacts)
}

func (g *group) Bullseye() *brevity.Bullseye {
	if g.bullseye == nil {
		return nil
	}
	mp := orb.MultiPoint{}
	for _, tf := range g.contacts {
		mp = append(mp, tf.Track.Front().Point)
	}
	center := mp.Bound().Center()

	bearing := unit.Angle(geo.Bearing(*g.bullseye, center)) * unit.Degree
	distance := unit.Length(geo.Distance(*g.bullseye, center)) * unit.Meter
	return brevity.NewBullseye(bearing, distance)
}

func (g *group) Altitude() unit.Length {
	var sum unit.Length
	for _, tf := range g.contacts {
		sum += tf.Track.Front().Altitude
	}
	mean := sum / unit.Length(len(g.contacts))
	rounded := unit.Length((math.Round(mean.Feet()/1000) * 1000)) * unit.Foot
	return rounded
}

func (g *group) Weeds() bool {
	return g.Altitude() < 1000*unit.Foot
}

func (g *group) Track() brevity.Track {
	if len(g.contacts) == 0 {
		return brevity.UnknownDirection
	}
	return g.contacts[0].Direction()
}

func (g *group) TrackAngle() unit.Angle {
	// TODO interpolate from all members
	return g.contacts[0].LastKnown().Heading
}

func (g *group) Aspect() brevity.Aspect {
	if g.aspect == nil {
		return brevity.UnknownAspect
	}
	return *g.aspect
}

func (g *group) SetAspect(aspect *brevity.Aspect) {
	g.aspect = aspect
}

func (g *group) BRAA() brevity.BRAA {
	return g.braa
}

func (g *group) Declaration() brevity.Declaration {
	return g.declaraction
}

func (g *group) SetDeclaration(declaration brevity.Declaration) {
	g.declaraction = declaration
}

func (g *group) Heavy() bool {
	return len(g.contacts) >= 3
}

func (g *group) Platforms() []string {
	return g.platforms
}

func (g *group) High() bool {
	return g.Altitude() > 40000*unit.Foot
}

func (g *group) Fast() bool {
	return false
}

func (g *group) VeryFast() bool {
	return false
}
