package types

// Coalition is the ID of a coalition in DCS World.
type Coalition int

const (
	// CoalitionRed is the ID of the red coalition.
	CoalitionRed = 1
	// CoalitionBlue is the ID of the blue coalition.
	CoalitionBlue = 2
	// CoalitionNeutrals is the ID of the neutral coalition.
	CoalitionNeutrals = 3
)

// IsSpectator returns true if the given coalition is not red or blue. SRS considers any other coalition ID to be a spectator.
func IsSpectator(c Coalition) bool {
	return (c != CoalitionRed) && (c != CoalitionBlue)
}
