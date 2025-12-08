package radar

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/encyclopedia"
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/rs/zerolog/log"

	"github.com/paulmach/orb/geo"
)

type group struct {
	isThreat    bool
	contacts    []*trackfiles.Trackfile
	bullseye    *orb.Point
	braa        brevity.BRAA
	aspect      *brevity.Aspect
	declaration brevity.Declaration
	mergedWith  int
}

var _ brevity.Group = &group{}

// Threat implements [brevity.Group.Threat].
func (g *group) Threat() bool {
	return g.isThreat
}

// SetThreat implements [brevity.Group.SetThreat].
func (g *group) SetThreat(isThreat bool) {
	g.isThreat = isThreat
}

// Contacts implements [brevity.Group.Contacts].
func (g *group) Contacts() int {
	return len(g.contacts)
}

// Bullseye implements [brevity.Group.Bullseye].
func (g *group) Bullseye() *brevity.Bullseye {
	if g.bullseye == nil {
		return nil
	}

	declination, err := bearings.Declination(*g.bullseye, g.missionTime())

	if err != nil {
		log.Error().Err(err).Stringer("group", g).Msg("failed to get declination for group")
	}
	point := g.point()
	bearing := spatial.TrueBearing(*g.bullseye, point).Magnetic(declination)
	distance := spatial.Distance(*g.bullseye, point)
	return brevity.NewBullseye(bearing, distance)
}

func (g *group) Stacks() []brevity.Stack {
	altitudes := []unit.Length{}
	for _, trackfile := range g.contacts {
		altitudes = append(altitudes, trackfile.LastKnown().Altitude)
	}
	return brevity.Stacks(altitudes...)
}

func (g *group) Altitude() unit.Length {
	stacks := g.Stacks()
	if len(stacks) == 0 {
		return 0
	}
	return stacks[0].Altitude
}

func (g *group) altitudes() []unit.Length {
	stacks := g.Stacks()
	altitudes := make([]unit.Length, 0, len(stacks))
	for _, stack := range stacks {
		altitudes = append(altitudes, stack.Altitude)
	}
	return altitudes
}

// Track implements [brevity.Group.Track].
func (g *group) Track() brevity.Track {
	if len(g.contacts) == 0 || g.Declaration() == brevity.Furball {
		return brevity.UnknownDirection
	}
	// TODO interpolate from all members
	return g.contacts[0].Direction()
}

func (g *group) course() bearings.Bearing {
	// TODO interpolate from all members
	return g.contacts[0].Course()
}

// Aspect implements [brevity.Group.Aspect].
func (g *group) Aspect() brevity.Aspect {
	if g.aspect == nil || g.Declaration() == brevity.Furball {
		return brevity.UnknownAspect
	}
	return *g.aspect
}

// SetAspect implements [brevity.Group.SetAspect].
func (g *group) SetAspect(aspect *brevity.Aspect) {
	g.aspect = aspect
}

// BRAA implements [brevity.Group.BRAA].
func (g *group) BRAA() brevity.BRAA {
	return g.braa
}

// Declaration implements [brevity.Group.Declaration].
func (g *group) Declaration() brevity.Declaration {
	return g.declaration
}

// SetDeclaration implements [brevity.Group.SetDeclaration].
func (g *group) SetDeclaration(declaration brevity.Declaration) {
	g.declaration = declaration
}

// Heavy implements [brevity.Group.Heavy].
func (g *group) Heavy() bool {
	return len(g.contacts) >= 3
}

// Platforms implements [brevity.Group.Platforms].
func (g *group) Platforms() []string {
	platforms := make(map[string]struct{})
	for _, trackfile := range g.contacts {
		var name string
		data, ok := encyclopedia.GetAircraftData(trackfile.Contact.ACMIName)
		if ok {
			for _, reportingName := range []string{data.NATOReportingName, data.Nickname, data.OfficialName, data.PlatformDesignation} {
				if reportingName != "" {
					name = reportingName
					break
				}
			}
		}
		platforms[name] = struct{}{}
	}
	result := make([]string, 0, len(platforms))
	for platform := range platforms {
		result = append(result, platform)
	}
	return result
}

// High implements [brevity.Group.High].
func (g *group) High() bool {
	return g.Altitude() > 40000*unit.Foot
}

func (g *group) speed() unit.Speed {
	speed := unit.Speed(0)
	for _, trackfile := range g.contacts {
		if trackfile.Speed() > speed {
			speed = trackfile.Speed()
		}
	}
	return speed
}

// Fast implements [brevity.Group.Fast].
func (g *group) Fast() bool {
	return g.speed() > 600*unit.Knot && !g.VeryFast()
}

// VeryFast implements [brevity.Group.VeryFast].
func (g *group) VeryFast() bool {
	return g.speed() > 900*unit.Knot
}

// MergedWith implements [brevity.Group.MergedWith].
func (g *group) MergedWith() int {
	return g.mergedWith
}

// SetMergedWith implements [brevity.Group.SetMergedWith].
func (g *group) SetMergedWith(mergedWith int) {
	g.mergedWith = mergedWith
}

func (g *group) String() string {
	location := ""
	if g.braa != nil {
		location = fmt.Sprintf(
			"BRAA %d %d %d %s",
			int(g.BRAA().Bearing().RoundedDegrees()),
			int(g.BRAA().Range().NauticalMiles()),
			int(g.BRAA().Altitude().Feet()),
			g.BRAA().Aspect(),
		)
	} else if g.bullseye != nil {
		location = fmt.Sprintf(
			"BULLSEYE %d/%d",
			int(g.Bullseye().Bearing().RoundedDegrees()),
			int(g.Bullseye().Distance().NauticalMiles()),
		)
	}

	s := fmt.Sprintf(
		"%s %s (%d) %s",
		location,
		g.Declaration(),
		g.Contacts(),
		strings.Join(g.Platforms(), ","),
	)
	if g.isThreat {
		s = "THREAT " + s
	}
	return s
}

// category of the group.
func (g *group) category() brevity.ContactCategory {
	aircraft, ok := encyclopedia.GetAircraftData(g.contacts[0].Contact.ACMIName)
	if !ok {
		// GUESS LOL
		return brevity.FixedWing
	}
	return aircraft.Category()
}

func (g *group) isArmed() bool {
	var hasData, result bool
	for _, trackfile := range g.contacts {
		if data, ok := encyclopedia.GetAircraftData(trackfile.Contact.ACMIName); ok {
			hasData = true
			if !data.HasTag(encyclopedia.Unarmed) {
				result = true
				break
			}
		}
	}
	if !hasData {
		result = true // Assumed armed if aircraft not in encyclopedia
	}
	return result
}

func (g *group) isFighter() bool {
	result := false
	for _, trackfile := range g.contacts {
		data, ok := encyclopedia.GetAircraftData(trackfile.Contact.ACMIName)
		if ok && data.HasTag(encyclopedia.Fighter) {
			result = true
			break
		}
	}
	return result
}

// point returns the center point of the group.
func (g *group) point() orb.Point {
	center := g.contacts[0].LastKnown().Point
	for _, trackfile := range g.contacts[1:] {
		center = geo.Midpoint(center, trackfile.LastKnown().Point)
	}
	return center
}

// missionTime returns the mission-time timestamp of the most recent trackfile in the group.
func (g *group) missionTime() time.Time {
	var latest time.Time
	for _, trackfile := range g.contacts {
		if trackfile.LastKnown().Time.After(latest) {
			latest = trackfile.LastKnown().Time
		}
	}
	return latest
}

// threatRadius returns the highest threat radius of all contacts in the group.
func (g *group) threatRadius() unit.Length {
	highest := unit.Length(0)
	for _, trackfile := range g.contacts {
		radius := encyclopedia.SAR2AR1Threat
		if data, ok := encyclopedia.GetAircraftData(trackfile.Contact.ACMIName); ok {
			radius = data.ThreatRadius()
		}
		if radius > highest {
			highest = radius
		}
	}
	return highest
}

func (g *group) ObjectIDs() []uint64 {
	ids := make([]uint64, 0, len(g.contacts))
	for _, trackfile := range g.contacts {
		ids = append(ids, trackfile.Contact.ID)
	}
	slices.Sort(ids)
	return ids
}
