package brevity

import "github.com/martinlindhe/unit"

// ThreatCall reports that a fighter will piece the threat range to a friendly aircraft.
// THREAT is more complicated in the real world, so this bot offers a simplified version.
// Reference: ATP 3-52.4 Chapter V section 18
type ThreatCall struct {
	// Callsign of the friendly aircraft under threat.
	Callsign string
	// Group that is threatening the friendly aircraft.
	Group Group
}

// MandatoryThreatDistance is the distance at which a contact is considered a threat regardless of aspect.
// Reference: ATP 3-52.4 Chapter V section 18 subsection c
const MandatoryThreatDistance = 35 * unit.NauticalMile
