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
		return NaturalLanguageResponse{
			Subtitle: fmt.Sprintf(
				"%s, %s, contact, alpha check bullseye %s/%d",
				c.composeCallsigns(response.Callsign),
				c.composeCallsigns(c.Callsign),
				response.Location.Bearing().String(),
				int(response.Location.Distance().NauticalMiles()),
			),
			Speech: fmt.Sprintf(
				"%s, %s, contact, alpha check bullseye %s, %d",
				c.composeCallsigns(response.Callsign),
				c.composeCallsigns(c.Callsign),
				pronounceBearing(response.Location.Bearing()),
				int(response.Location.Distance().NauticalMiles()),
			),
		}
	}

	reply := response.Callsign + ", negative contact."
	return NaturalLanguageResponse{
		Subtitle: reply,
		Speech:   reply,
	}
}
