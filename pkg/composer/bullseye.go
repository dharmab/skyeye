package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeBullseye constructs natural language brevity for communicating Bullseye information.
// Example: "bullseye 270/20"
func (c *composer) ComposeBullseye(bullseye brevity.Bullseye) NaturalLanguageResponse {
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf(
			"bullseye %d/%d",
			int(bullseye.Bearing().Degrees()),
			int(bullseye.Distance().NauticalMiles()),
		),
		Speech: fmt.Sprintf(
			"bullseye %s, %d",
			PronounceBearing(int(bullseye.Bearing().Degrees())),
			int(bullseye.Distance().NauticalMiles()),
		),
	}
}
