package composer

import (
	"fmt"
	"math/rand/v2"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeNegativeRadarContactResponse constructs natural language brevity for saying the controller cannot find a contact on the radar.
func (c *Composer) ComposeNegativeRadarContactResponse(response brevity.NegativeRadarContactResponse) NaturalLanguageResponse {
	replies := []string{
		"%s, negative radar contact. Double check your callsign.",
		"%s, negative radar contact. Check your callsign.",
		"%s, negative radar contact. Verify your callsign.",
		"%s, negative radar contact. Confirm your callsign.",
		"%s, negative radar contact. Send it again for me.",
		"%s, negative radar contact. I might have misheard your callsign.",
		"%s, negative radar contact. Is that the right callsign?",
		"%s, negative radar contact. Possible I misheard the callsign.",
		"%s, negative radar contact. No contact with that callsign on scope.",
		"%s, negative radar contact. Can't find that callsign on scope.",
		"%s, negative radar contact. I don't see that callsign on scope.",
		"%s, negative radar contact. I don't have that callsign on scope.",
		"%s, negative radar contact. I do not have that callsign on scope.",
	}
	reply := fmt.Sprintf(replies[rand.IntN(len(replies))], c.composeCallsigns(response.Callsign))
	return NaturalLanguageResponse{
		Subtitle: reply,
		Speech:   reply,
	}
}
