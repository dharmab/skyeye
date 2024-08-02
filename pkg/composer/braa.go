package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// ComposeBRAA constructs natural language brevity for communicating BRAA information.
// Example: "BRAA 270/20, 20000, hot"
func (c *composer) ComposeBRAA(braa brevity.BRAA) NaturalLanguageResponse {
	if !braa.Bearing().IsMagnetic() {
		log.Error().Any("bearing", braa.Bearing()).Msg("bearing provided to ComposeBRAA should be magnetic")
	}
	var aspect string
	if braa.Aspect() != brevity.UnknownAspect {
		aspect = string(braa.Aspect())
	}
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf(
			"BRAA %s/%d, %d, %s",
			braa.Bearing().String(),
			int(braa.Range().NauticalMiles()),
			int(braa.Altitude().Feet()),
			aspect,
		),
		Speech: fmt.Sprintf(
			"brah %s, %d, %d, %s",
			PronounceBearing(braa.Bearing()),
			int(braa.Range().NauticalMiles()),
			int(braa.Altitude().Feet()),
			aspect,
		),
	}
}
