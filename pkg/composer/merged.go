package composer

import (
	"github.com/dharmab/skyeye/pkg/brevity"
)

func (c *composer) ComposeMergedCall(call brevity.MergedCall) NaturalLanguageResponse {
	reply := c.ComposeCallsigns(call.Callsigns...) + ", merged."
	return NaturalLanguageResponse{
		Subtitle: reply,
		Speech:   reply,
	}
}
