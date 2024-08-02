package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeSayAgainResponse implements [Composer.ComposeSayAgainResponse].
func (c *composer) ComposeSayAgainResponse(response brevity.SayAgainResponse) NaturalLanguageResponse {
	if response.Callsign != "" {
		return NaturalLanguageResponse{
			Subtitle: fmt.Sprintf("%s, sorry, I didn't understand. Say again.", response.Callsign),
			Speech:   fmt.Sprintf("%s, sorry, I didn't understand. Say again.", response.Callsign),
		}
	}
	return NaturalLanguageResponse{
		Subtitle: "I heard my callsign, but I did not understand the request. Say again.",
		Speech:   "I heard my callsign, but I did not understand the request. Say again.",
	}
}
