package composer

import "github.com/dharmab/skyeye/pkg/brevity"

func (c *composer) ComposeAlphaCheckResponse(brevity.AlphaCheckResponse) NaturalLanguageResponse {
	return NaturalLanguageResponse{
		Subtitle: "ALPHA CHECK not yet implemented",
		Speech:   "Sorry, I don't know how to respond to ALPHA CHECK yet. I'm still learning!",
	}
}
