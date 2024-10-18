package brevity

import (
	"github.com/martinlindhe/unit"
)

type MergedCall struct {
	// Callsigns of the friendly aircraft in the merge.
	Callsigns []string
}

const (
	// MergeEntryDistance is the distance at which contacts are considered to enter the merge.
	MergeEntryDistance = 3 * unit.NauticalMile
	// MergeExitDistance is the distance at which contacts are considered to exit the merge.
	MergeExitDistance = 5 * unit.NauticalMile
)
