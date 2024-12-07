package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// composeBullseye constructs natural language brevity for communicating Bullseye information.
func (*Composer) composeBullseye(bullseye brevity.Bullseye) NaturalLanguageResponse {
	if !bullseye.Bearing().IsMagnetic() {
		log.Error().Stringer("bearing", bullseye.Bearing()).Msg("bearing provided to ComposeBullseye should be magnetic")
	}
	if bullseye.Distance().NauticalMiles() <= 5 {
		return NaturalLanguageResponse{
			Subtitle: "at bullseye",
			Speech:   "at bullseye",
		}
	}
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf(
			"bullseye %s/%d",
			bullseye.Bearing().String(),
			int(bullseye.Distance().NauticalMiles()),
		),
		Speech: fmt.Sprintf(
			"bullseye %s, %d",
			pronounceBearing(bullseye.Bearing()),
			int(bullseye.Distance().NauticalMiles()),
		),
	}
}
