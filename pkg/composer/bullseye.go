package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// ComposeBullseye constructs natural language brevity for communicating Bullseye information.
// Example: "bullseye 270/20"
func (c *composer) ComposeBullseye(bullseye brevity.Bullseye) NaturalLanguageResponse {
	if !bullseye.Bearing().IsMagnetic() {
		log.Error().Any("bearing", bullseye.Bearing()).Msg("bearing provided to ComposeBullseye should be magnetic")
	}
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf(
			"bullseye %s/%d",
			bullseye.Bearing().String(),
			int(bullseye.Distance().NauticalMiles()),
		),
		Speech: fmt.Sprintf(
			"bullseye %s, %d",
			PronounceBearing(bullseye.Bearing()),
			int(bullseye.Distance().NauticalMiles()),
		),
	}
}
