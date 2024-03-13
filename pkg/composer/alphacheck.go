package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func (c *composer) ComposeAlphaCheckResponse(r brevity.AlphaCheckResponse) NaturalLanguageResponse {
	if r.Status() {
		return NaturalLanguageResponse{
			Subtitle: fmt.Sprintf(
				"%s, %s, contact, alpha check bullseye %d/%d",
				r.Callsign(),
				c.callsign,
				int(r.Location().Bearing().Degrees()),
				int(r.Location().Distance().NauticalMiles()),
			),
			Speech: fmt.Sprintf(
				"%s, %s, contact, alpha check bullseye %s, %d",
				r.Callsign(),
				c.callsign,
				PronounceInt(int(r.Location().Bearing().Degrees())),
				int(r.Location().Distance().NauticalMiles()),
			),
		}
	}

	reply := fmt.Sprintf("%s, negative contact", r.Callsign())
	return NaturalLanguageResponse{
		Subtitle: reply,
		Speech:   reply,
	}
}
