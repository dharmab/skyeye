package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeAlphaCheckResponse implements [Composer.ComposeAlphaCheckResponse].
func (c *composer) ComposeAlphaCheckResponse(response brevity.AlphaCheckResponse) NaturalLanguageResponse {
	if response.Status {
		return NaturalLanguageResponse{
			Subtitle: fmt.Sprintf(
				"%s, %s, contact, alpha check bullseye %d/%d",
				response.Callsign,
				c.callsign,
				int(response.Location.Bearing().Degrees()),
				int(response.Location.Distance().NauticalMiles()),
			),
			Speech: fmt.Sprintf(
				"%s, %s, contact, alpha check bullseye %s, %d",
				response.Callsign,
				c.callsign,
				PronounceBearing(int(response.Location.Bearing().Degrees())),
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
