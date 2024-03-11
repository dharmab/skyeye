package composer

import "github.com/dharmab/skyeye/pkg/brevity"

func (c *composer) ComposeFadedCall(brevity.FadedCall) NaturalLanguageResponse {
	return NaturalLanguageResponse{
		Subtitle: "FADED not yet implemented",
		Speech:   "Sorry, I don't know how make a FADED call yet. I'm still learning!",
	}
}
