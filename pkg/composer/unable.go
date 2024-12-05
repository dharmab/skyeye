package composer

import (
	"fmt"
	"math/rand/v2"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeSayAgainResponse constructs natural language brevity for asking a caller to repeat their last transmission.
func (c *Composer) ComposeSayAgainResponse(response brevity.SayAgainResponse) NaturalLanguageResponse {
	replies := map[bool][]string{
		true: {
			"%s, sorry, I didn't understand. Say again.",
			"%s, I didn't catch that. Say again.",
			"%s, I didn't understand. Say again.",
			"%s, say again.",
			"%s, I didn't get that. Say again.",
			"%s, I only got the first part of that. Say again.",
		},
		false: {
			"I heard my callsign, but I did not understand the request. Say again.",
			"I heard someone call me, but I didn't understand what they said. Say again.",
			"I only got the first part of that. Say again.",
			"Sorry, I only caught part of that. Say again.",
		},
	}
	haveCallsign := response.Callsign != ""
	variation := replies[haveCallsign][rand.IntN(len(replies[haveCallsign]))]
	reply := ""
	if haveCallsign {
		reply = fmt.Sprintf(variation, c.composeCallsigns(response.Callsign))
	} else {
		reply = variation
	}
	return NaturalLanguageResponse{
		Subtitle: reply,
		Speech:   reply,
	}
}
