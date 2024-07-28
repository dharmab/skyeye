package radar

import (
	"math"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/encyclopedia"
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

// Threat implements [brevity.Group.Threat]
func (g *group) Threat() bool {
	return g.isThreat
}

// SetThreat implements [brevity.Group.SetThreat]
func (g *group) SetThreat(isThreat bool) {
	g.isThreat = isThreat
}

// Contacts implements [brevity.Group.Contacts]
func (g *group) Contacts() int {
	return len(g.contacts)
}

// Bullseye implements [brevity.Group.Bullseye]
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

// Altitude implements [brevity.Group.Altitude] by averaging the altitudes of all contacts in the group
func (g *group) Altitude() unit.Length {
	var sum unit.Length
	for _, tf := range g.contacts {
		sum += tf.Track.Front().Altitude
	}
	mean := sum / unit.Length(len(g.contacts))
	rounded := unit.Length((math.Round(mean.Feet()/1000) * 1000)) * unit.Foot
	return rounded
}

// Weeds implements [brevity.Group.Weeds]
func (g *group) Weeds() bool {
	// TODO use AGL instead of MSL
	return g.Altitude() < 1000*unit.Foot
}

// Track implements [brevity.Group.Track]
func (g *group) Track() brevity.Track {
	if len(g.contacts) == 0 {
		return brevity.UnknownDirection
	}
	// TODO interpolate from all members
	return g.contacts[0].Direction()
}

// Aspect implements [brevity.Group.Aspect]
func (g *group) Aspect() brevity.Aspect {
	if g.aspect == nil {
		return brevity.UnknownAspect
	}
	return *g.aspect
}

// SetAspect implements [brevity.Group.SetAspect]
func (g *group) SetAspect(aspect *brevity.Aspect) {
	g.aspect = aspect
}

// BRAA implements [brevity.Group.BRAA]
func (g *group) BRAA() brevity.BRAA {
	return g.braa
}

// Declaration implements [brevity.Group.Declaration]
func (g *group) Declaration() brevity.Declaration {
	return g.declaraction
}

// SetDeclaration implements [brevity.Group.SetDeclaration]
func (g *group) SetDeclaration(declaration brevity.Declaration) {
	g.declaraction = declaration
}

// Heavy implements [brevity.Group.Heavy]
func (g *group) Heavy() bool {
	return len(g.contacts) >= 3
}

// Platforms implements [brevity.Group.Platforms]
func (g *group) Platforms() []string {
	return g.platforms
}

// High implements [brevity.Group.High]
func (g *group) High() bool {
	return g.Altitude() > 40000*unit.Foot
}

// Fast implements [brevity.Group.Fast]
func (g *group) Fast() bool {
	return false
}

// VeryFast implements [brevity.Group.VeryFast]
func (g *group) VeryFast() bool {
	return false
}

// category of the group
func (g *group) category() brevity.ContactCategory {
	aircraft, ok := encyclopedia.GetAircraftData(g.contacts[0].Contact.ACMIName)
	if !ok {
		// GUESS LOL
		return brevity.FixedWing
	}
	return aircraft.Category()
}

// point returns the center point of the group
func (g *group) point() orb.Point {
	center := g.contacts[0].LastKnown().Point
	for i := 1; i < len(g.contacts); i++ {
		center = geo.Midpoint(center, g.contacts[i].LastKnown().Point)
	}
	return center
}

// threatClass returns the highest threat class of all contacts in the group.
func (g *group) threatClass() encyclopedia.ThreatClass {
	var groupThreatClass encyclopedia.ThreatClass
	for _, tf := range g.contacts {
		data, ok := encyclopedia.GetAircraftData(tf.Contact.ACMIName)
		if !ok {
			continue
		}
		contactThreatClass := data.ThreatClass()
		if contactThreatClass > groupThreatClass {
			groupThreatClass = contactThreatClass
		}
	}
	return groupThreatClass
}
