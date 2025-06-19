package composer

import (
	"fmt"
	"math/rand/v2"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeNegativeRadarContactResponse constructs natural language brevity for saying the controller cannot find a contact on the radar.
func (c *Composer) ComposeNegativeRadarContactResponse(response brevity.NegativeRadarContactResponse) NaturalLanguageResponse {
	prefix := "%s, negative radar contact. "
	suffixes := []string{
		"Double check your callsign.",
		"Check your callsign.",
		"Verify your callsign.",
		"Confirm your callsign.",
		"Send it again for me.",
		"I might have misheard your callsign.",
		"Is that the right callsign?",
		"Possible I misheard the callsign.",
		"No contact with that callsign on scope.",
		"Can't find that callsign on scope.",
		"I don't see that callsign on scope.",
		"I don't have that callsign on scope.",
		"I do not have that callsign on scope.",
	}
	format := prefix + suffixes[rand.IntN(len(suffixes))]
	reply := fmt.Sprintf(format, c.composeCallsigns(response.Callsign))
	return NaturalLanguageResponse{
		Subtitle: reply,
		Speech:   reply,
	}
}
