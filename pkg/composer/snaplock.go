package composer

import "github.com/dharmab/skyeye/pkg/brevity"

func (c *composer) ComposeSnaplockResponse(brevity.SnaplockResponse) NaturalLanguageResponse {
	return NaturalLanguageResponse{
		Subtitle: "SNAP LOCK not yet implemented",
		Speech:   "Sorry, I don't know how to respond to SNAP LOCK yet. I'm still learning!",
	}
}
