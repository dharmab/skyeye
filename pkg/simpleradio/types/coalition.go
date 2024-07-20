package types

import (
	"github.com/dharmab/skyeye/pkg/coalitions"
)

// IsSpectator returns true if the given coalition is not red or blue. SRS considers any other coalition ID to be a spectator.
func IsSpectator(c coalitions.Coalition) bool {
	return (c != coalitions.Red) && (c != coalitions.Blue)
}
