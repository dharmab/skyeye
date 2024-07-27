package brevity

import (
	"github.com/martinlindhe/unit"
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

func TrackFromBearing(bearing unit.Angle) Track {
	θ := int(bearing.Degrees()) % 360
	log.Debug().Int("bearing", θ).Msg("TrackFromBearing")
	switch {
	case θ > 337 || θ <= 22:
		return North
	case θ > 22 && θ <= 67:
		return Northeast
	case θ > 67 && θ <= 112:
		return East
	case θ > 112 && θ <= 157:
		return Southeast
	case θ > 157 && θ <= 202:
		return South
	case θ > 202 && θ <= 247:
		return Southwest
	case θ > 247 && θ <= 292:
		return West
	case θ > 292 && θ <= 337:
		return Northwest
	default:
		return UnknownDirection
	}
}
