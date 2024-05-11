package brevity

import "github.com/martinlindhe/unit"

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
	θ := bearing.Degrees()
	switch {
	case θ >= 337.5 || θ < 22.5:
		return North
	case θ >= 22.5 && θ < 67.5:
		return Northeast
	case θ >= 67.5 && θ < 112.5:
		return East
	case θ >= 112.5 && θ < 157.5:
		return Southeast
	case θ >= 157.5 && θ < 202.5:
		return South
	case θ >= 202.5 && θ < 247.5:
		return Southwest
	case θ >= 247.5 && θ < 292.5:
		return West
	case θ >= 292.5 && θ < 337.5:
		return Northwest
	default:
		return UnknownDirection
	}
}
