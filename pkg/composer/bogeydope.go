package composer

import "github.com/dharmab/skyeye/pkg/brevity"

func (c *composer) ComposeBogeyDopeResponse(brevity.BogeyDopeResponse) NaturalLanguageResponse {
	return NaturalLanguageResponse{
		Subtitle: "BOGEY DOPE not yet implemented",
		Speech:   "Sorry, I don't know how to respond to BOGEY DOPE yet. I'm still learning!",
	}
}
