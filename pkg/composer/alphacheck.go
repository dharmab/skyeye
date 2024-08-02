package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// ComposeAlphaCheckResponse implements [Composer.ComposeAlphaCheckResponse].
func (c *composer) ComposeAlphaCheckResponse(response brevity.AlphaCheckResponse) NaturalLanguageResponse {
	if response.Status {
		if !response.Location.Bearing().IsMagnetic() {
			log.Error().Any("bearing", response.Location.Bearing()).Msg("bearing provided to ComposeAlphaCheckResponse should be magnetic")
		}
		return NaturalLanguageResponse{
			Subtitle: fmt.Sprintf(
				"%s, %s, contact, alpha check bullseye %s/%d",
				response.Callsign,
				c.callsign,
				response.Location.Bearing().String(),
				int(response.Location.Distance().NauticalMiles()),
			),
			Speech: fmt.Sprintf(
				"%s, %s, contact, alpha check bullseye %s, %d",
				response.Callsign,
				c.callsign,
				PronounceBearing(response.Location.Bearing()),
				int(response.Location.Distance().NauticalMiles()),
			),
		}
	}

	reply := fmt.Sprintf("%s, negative contact", response.Callsign)
	return NaturalLanguageResponse{
		Subtitle: reply,
		Speech:   reply,
	}
}
