package brevity

// Track is a compass direction.
type Track int

const (
	UnknownDirection Track = -1
	North            Track = iota
	Northeast
	East
	Southeast
	South
	Southwest
	West
	Northwest
)
