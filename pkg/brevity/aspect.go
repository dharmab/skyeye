package brevity

import (
	"math"

	"github.com/martinlindhe/unit"
)

// Aspect indicates the target aspect or aspect angle between a contact and fighter.
// Reference: ATP 3-52.4 Chapter IV section 6, Figure 1
type Aspect string

const (
	UnknownAspect Aspect = "maneuver"
	// Hot aspect is 0-30° target aspect or 180-150° aspect angle.
	Hot = "hot"
	// Flank is 40-70° target aspect or 140-110° aspect angle.
	Flank = "flank"
	// Beam is 80-110° target aspect or 100-70° aspect angle.
	Beam = "beam"
	// Drag is 120-180° target aspect or 60-0° aspect angle.
	Drag = "drag"
)

// AspectFromAngle computes target aspect based on the bearing from an aircraft to the target and the track direction of the target.
func AspectFromAngle(bearing unit.Angle, track unit.Angle) Aspect {
	targetBearing := unit.Angle(bearing.Degrees()+180) * unit.Degree
	targetAspect := int(math.Abs(track.Degrees()-targetBearing.Degrees())) % 360
	switch {
	case targetAspect <= 30:
		return Hot
	case targetAspect <= 70:
		return Flank
	case targetAspect <= 110:
		return Beam
	case targetAspect <= 180:
		return Drag
	default:
		return UnknownAspect
	}
}
