package locations

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/paulmach/orb"
	"gopkg.in/yaml.v3"
)

// ReservedNames are location names that cannot be used as custom location names.
var ReservedNames = []string{"tanker", "bullseye"}

// Location is a named geographic location that can be referenced in ALPHA CHECK and VECTOR requests.
type Location struct {
	Names     []string `json:"names" yaml:"names"`
	Longitude float64  `json:"longitude" yaml:"longitude"`
	Latitude  float64  `json:"latitude" yaml:"latitude"`
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

// LoadLocations parses location data from JSON or YAML. It tries JSON first, then falls back to YAML.
func LoadLocations(data []byte) ([]Location, error) {
	var locs []Location
	if err := json.Unmarshal(data, &locs); err != nil {
		locs = nil
		if yamlErr := yaml.Unmarshal(data, &locs); yamlErr != nil {
			return nil, fmt.Errorf("failed to parse locations as JSON or YAML: json: %w, yaml: %w", err, yamlErr)
		}
	}
	for _, loc := range locs {
		if err := loc.Validate(); err != nil {
			return nil, err
		}
	}
	return locs, nil
}
