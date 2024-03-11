package composer

import "github.com/dharmab/skyeye/pkg/brevity"

func (c *composer) ComposeRadioCheckResponse(brevity.RadioCheckResponse) NaturalLanguageResponse {
	return NaturalLanguageResponse{
		Subtitle: "RADIO CHECK not yet implemented",
		Speech:   "Sorry, I don't know how to respond to RADIO CHECK yet. I'm still learning!",
	}
}
