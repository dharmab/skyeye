package composer

import (
	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeSpikedResponse constructs natural language brevity for responding to a SPIKED call.
func (c *Composer) ComposeSpikedResponse(response brevity.SpikedResponseV2) NaturalLanguageResponse {
	return c.composeCorrelation("spike", response.Callsign, response.Status, response.Bearing, response.Group)
}
