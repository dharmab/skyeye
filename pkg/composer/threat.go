package composer

import "github.com/dharmab/skyeye/pkg/brevity"

func (c *composer) ComposeThreatCall(brevity.ThreatCall) NaturalLanguageResponse {
	return NaturalLanguageResponse{
		Subtitle: "THREAT not yet implemented",
		Speech:   "Sorry, I don't know how to make a THREAT call yet. I'm still learning!",
	}
}
