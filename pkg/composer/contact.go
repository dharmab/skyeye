package composer

import (
	"fmt"
	"math/rand"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeNegativeRadarContactResponse implements [Composer.ComposeNegativeRadarContactResponse].
func (c *composer) ComposeNegativeRadarContactResponse(r brevity.NegativeRadarContactResponse) NaturalLanguageResponse {
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
	s := fmt.Sprintf(replies[rand.Intn(len(replies))], r.Callsign)
	return NaturalLanguageResponse{
		Subtitle: s,
		Speech:   s,
	}
}
