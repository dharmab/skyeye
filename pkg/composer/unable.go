package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeSayAgainResponse implements [Composer.ComposeSayAgainResponse].
func (c *composer) ComposeSayAgainResponse(r brevity.SayAgainResponse) NaturalLanguageResponse {
	if r.Callsign != "" {
		return NaturalLanguageResponse{
			Subtitle: fmt.Sprintf("%s, sorry, I didn't understand. Say again.", r.Callsign),
			Speech:   fmt.Sprintf("%s, sorry, I didn't understand. Say again.", r.Callsign),
		}
	}
	return NaturalLanguageResponse{
		Subtitle: "I heard my callsign, but I did not understand the request. Say again.",
		Speech:   "I heard my callsign, but I did not understand the request. Say again.",
	}
}
