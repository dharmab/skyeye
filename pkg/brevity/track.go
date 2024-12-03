package brevity

import (
	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/rs/zerolog/log"
)

// Track is a compass direction.
type Track string

const (
	UnknownDirection Track = "unknown"
	North            Track = "north"
	Northeast        Track = "northeast"
	East             Track = "east"
	Southeast        Track = "southeast"
	South            Track = "south"
	Southwest        Track = "southwest"
	West             Track = "west"
	Northwest        Track = "northwest"
)

// TrackFromBearing computes a track direction based on the given magnetic bearing.
func TrackFromBearing(bearing bearings.Bearing) Track {
	if !bearing.IsMagnetic() {
		log.Warn().Stringer("bearing", bearing).Msg("bearing provided to TrackFromBearing should be magnetic")
	}
	θ := bearing.Degrees()
	switch {
	case θ >= 337.5 || θ < 22.5:
		return North
	case θ < 67.5:
		return Northeast
	case θ < 112.5:
		return East
	case θ < 157.5:
		return Southeast
	case θ < 202.5:
		return South
	case θ < 247.5:
		return Southwest
	case θ < 292.5:
		return West
	case θ < 337.5:
		return Northwest
	default:
		return UnknownDirection
	}
}

func (t Track) String() string {
	return string(t)
}
