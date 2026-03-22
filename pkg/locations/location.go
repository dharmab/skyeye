package locations

import (
	"fmt"
	"strings"

	"github.com/paulmach/orb"
)

// ReservedNames are location names that cannot be used as custom location names.
var ReservedNames = []string{"tanker", "bullseye"}

// Location is a named geographic location that can be referenced in ALPHA CHECK and VECTOR requests.
type Location struct {
	Names     []string `json:"names"`
	Longitude float64  `json:"longitude"`
	Latitude  float64  `json:"latitude"`
}

// Point returns the location as an orb.Point.
func (l Location) Point() orb.Point {
	return orb.Point{l.Longitude, l.Latitude}
}

// Validate checks that the location does not use any reserved names.
func (l Location) Validate() error {
	for _, name := range l.Names {
		for _, reserved := range ReservedNames {
			if strings.EqualFold(name, reserved) {
				return fmt.Errorf("location name %q is reserved and cannot be used as a custom location name", name)
			}
		}
	}
	return nil
}
