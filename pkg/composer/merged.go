package composer

import (
	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeMergedCall constructs natural language brevity for announcing a merge.
func (c *Composer) ComposeMergedCall(call brevity.MergedCall) NaturalLanguageResponse {
	reply := c.composeCallsigns(call.Callsigns...) + ", merged."
	return NaturalLanguageResponse{
		Subtitle: reply,
		Speech:   reply,
	}
}
