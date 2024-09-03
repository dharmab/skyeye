package brevity

import (
	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
)

// Aspect indicates the target aspect or aspect angle between a contact and fighter.
// Reference: ATP 3-52.4 Chapter IV section 6, Figure 1.
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

// AspectFromAngle computes target aspect based on the magnetic bearing from an aircraft to the target and the track direction of the target.
func AspectFromAngle(bearing bearings.Bearing, track bearings.Bearing) Aspect {
	if !bearing.IsMagnetic() || !track.IsMagnetic() {
		log.Warn().Stringer("bearing", bearing).Stringer("track", track).Msg("bearing and track provided to AspectFromAngle should be magnetic")
	}

	var θ float64
	if bearing.Reciprocal().Value() > track.Value() {
		θ = bearing.Reciprocal().Degrees() - track.Degrees()
	} else {
		θ = track.Reciprocal().Degrees() - bearing.Degrees()
	}
	θ = bearings.NewMagneticBearing(unit.Angle(θ) * unit.Degree).Degrees()

	switch {
	case 0 <= θ && θ <= 35:
		return Hot
	case 35 < θ && θ <= 75:
		return Flank
	case 75 < θ && θ <= 115:
		return Beam
	case 115 < θ && θ <= 245:
		return Drag
	case 245 < θ && θ <= 285:
		return Beam
	case 285 < θ && θ <= 325:
		return Flank
	case 325 < θ && θ <= 360:
		return Hot
	default:
		return UnknownAspect
	}
}
