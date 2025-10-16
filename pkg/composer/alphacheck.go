package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// ComposeAlphaCheckResponse constructs natural language brevity for responding to an ALPHA CHECK.
func (c *Composer) ComposeAlphaCheckResponse(response brevity.AlphaCheckResponse) NaturalLanguageResponse {
	if response.Status {
		if !response.Location.Bearing().IsMagnetic() {
			log.Error().Stringer("bearing", response.Location.Bearing()).Msg("bearing provided to ComposeAlphaCheckResponse should be magnetic")
		}
		callerCallsign := c.composeCallsigns(response.Callsign)
		controllerCallsign := c.composeCallsigns(c.Callsign)
		_range := response.Location.Distance()
		return NaturalLanguageResponse{
			Subtitle: fmt.Sprintf(
				"%s, %s, contact, alpha check bullseye %s/%d",
				callerCallsign,
				controllerCallsign,
				response.Location.Bearing().String(),
				int(_range.NauticalMiles()),
			),
			Speech: fmt.Sprintf(
				"%s, %s, contact, alpha check bullseye %s, %d",
				callerCallsign,
				controllerCallsign,
				pronounceBearing(response.Location.Bearing()),
				int(_range.NauticalMiles()),
			),
		}
	}

	reply := response.Callsign + ", negative contact."
	return NaturalLanguageResponse{
		Subtitle: reply,
		Speech:   reply,
	}
}
