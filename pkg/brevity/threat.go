package brevity

// ThreatCall reports that a fighter will piece the threat range to a friendly aircraft.
// THREAT is more complicated in the real world, so this bot offers a simplified version.
// Reference: ATP 3-52.4 Chapter V section 18
type ThreatCall interface {
	// Callsign of the friendly aircraft under threat.
	Callsign() string
	// Group that is threatening the friendly aircraft.
	Group() Group
}
