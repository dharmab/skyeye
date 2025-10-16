package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
)

// composeBullseye constructs natural language brevity for communicating Bullseye information.
func (*Composer) composeBullseye(bullseye brevity.Bullseye) NaturalLanguageResponse {
	if !bullseye.Bearing().IsMagnetic() {
		log.Error().Stringer("bearing", bullseye.Bearing()).Msg("bearing provided to ComposeBullseye should be magnetic")
	}
	_range := bullseye.Distance()
	const bullseyeRadius = 5 * unit.NauticalMile
	if _range <= bullseyeRadius {
		return NaturalLanguageResponse{
			Subtitle: "at bullseye",
			Speech:   "at bullseye",
		}
	}
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf(
			"bullseye %s/%d",
			bullseye.Bearing().String(),
			int(_range.NauticalMiles()),
		),
		Speech: fmt.Sprintf(
			"bullseye %s, %d",
			pronounceBearing(bullseye.Bearing()),
			int(_range.NauticalMiles()),
		),
	}
}
