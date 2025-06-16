package composer

import (
	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeStrobeResponse constructs natural language brevity for responding to a STROBE call.
func (c *Composer) ComposeStrobeResponse(response brevity.StrobeResponse) NaturalLanguageResponse {
	return c.composeCorrelation("strobe", response.Callsign, response.Status, response.Bearing, response.Group)
}
