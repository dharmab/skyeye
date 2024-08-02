package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeThreatCall implements [Composer.ComposeThreatCall].
func (c *composer) ComposeThreatCall(call brevity.ThreatCall) NaturalLanguageResponse {
	group := c.ComposeGroup(call.Group)
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("%s, %s", call.Callsign, group.Subtitle),
		Speech:   fmt.Sprintf("%s, %s", call.Callsign, group.Speech),
	}
}
