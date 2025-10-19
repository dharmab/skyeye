package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeThreatCall constructs natural language brevity for announcing a threat.
func (c *Composer) ComposeThreatCall(call brevity.ThreatCall) NaturalLanguageResponse {
	group := c.composeGroup(call.Group)
	callsignList := c.composeCallsigns(call.Callsigns...)
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("%s, %s", callsignList, lowerFirst(group.Subtitle)),
		Speech:   fmt.Sprintf("%s, %s", callsignList, group.Speech),
	}
}
