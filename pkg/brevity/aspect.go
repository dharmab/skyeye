package brevity

import (
	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
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

// AspectFromAngle computes target aspect based on the magnetic bearing from an aircraft to the target and the track direction of the target.
func AspectFromAngle(bearing bearings.Bearing, track bearings.Bearing) Aspect {
	if !bearing.IsMagnetic() || !track.IsMagnetic() {
		log.Warn().Any("bearing", bearing).Any("track", track).Msg("bearing and track provided to AspectFromAngle should be magnetic")
	}

	var targetAspect unit.Angle
	if track.Value() > bearing.Value() {
		targetAspect = track.Value() - bearing.Magnetic(0).Value()
	} else {
		targetAspect = bearing.Value() - track.Value()
	}

	θ := targetAspect.Degrees()
	switch {
	case θ <= 30:
		return Hot
	case θ <= 70:
		return Flank
	case θ <= 110:
		return Beam
	case θ <= 180:
		return Drag
	default:
		return UnknownAspect
	}
}
