package locations

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/paulmach/orb"
	"gopkg.in/yaml.v3"
)

// ReservedNames are location names that cannot be used as custom location names.
var ReservedNames = []string{"tanker", "bullseye"}

// Location is a named geographic location that can be referenced in VECTOR requests.
type Location struct {
	Names     []string `json:"names" yaml:"names"`
	Longitude float64  `json:"longitude" yaml:"longitude"`
	Latitude  float64  `json:"latitude" yaml:"latitude"`
}

// Point returns the location as an orb.Point.
func (l Location) Point() orb.Point {
	return orb.Point{l.Longitude, l.Latitude}
}

// Validate checks that the location has at least one non-empty name, does
// not use any reserved names, and has coordinates within valid bounds.
func (l Location) Validate() error {
	if len(l.Names) == 0 {
		return errors.New("location must have at least one name")
	}
	for _, name := range l.Names {
		if strings.TrimSpace(name) == "" {
			return errors.New("location name must not be empty or whitespace")
		}
		for _, reserved := range ReservedNames {
			if strings.EqualFold(name, reserved) {
				return fmt.Errorf("location name %q is reserved and cannot be used as a custom location name", name)
			}
		}
	}
	if l.Latitude < -90 || l.Latitude > 90 {
		return fmt.Errorf("location latitude %v is outside the valid range [-90, 90]", l.Latitude)
	}
	if l.Longitude < -180 || l.Longitude > 180 {
		return fmt.Errorf("location longitude %v is outside the valid range [-180, 180]", l.Longitude)
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
