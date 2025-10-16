package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeSnaplockResponse constructs natural language brevity for responding to a SNAPLOCK call.
func (c *Composer) ComposeSnaplockResponse(response brevity.SnaplockResponse) NaturalLanguageResponse {
	if response.Declaration == brevity.Hostile || response.Declaration == brevity.Friendly {
		info := c.composeCoreInformationFormat(response.Group)
		callerCallsign := c.composeCallsigns(response.Callsign)
		return NaturalLanguageResponse{
			Subtitle: fmt.Sprintf("%s, %s", callerCallsign, info.Subtitle),
			Speech:   fmt.Sprintf("%s, %s", callerCallsign, info.Speech),
		}
	}

	reply := fmt.Sprintf("%s, %s", c.composeCallsigns(response.Callsign), response.Declaration)
	return NaturalLanguageResponse{
		Subtitle: reply,
		Speech:   reply,
	}
}
