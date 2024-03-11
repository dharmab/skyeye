package composer

import "github.com/dharmab/skyeye/pkg/brevity"

func (c *composer) ComposeSpikedResponse(brevity.SpikedResponse) NaturalLanguageResponse {
	return NaturalLanguageResponse{
		Subtitle: "SPIKED not yet implemented",
		Speech:   "Sorry, I don't know how to respond to SPIKED yet. I'm still learning!",
	}
}
