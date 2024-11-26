package composer

import (
	"math/rand/v2"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func (c *composer) ComposeCheckInResponse(response brevity.CheckInResponse) NaturalLanguageResponse {
	replies := []string{
		", I'm not sure if you wanted a radio check or an alpha check. Or did you just want to say hi?",
		", I'm not sure if you wanted a radio check or an alpha check. Or are you trying to flirt? In that case, if you find me in the officer's club later, you can buy me an old fashioned.",
		", did you want a radio check or an alpha check? Or are you just trying to get my attention? In that case, I'm flattered.",
		", did you want a radio check or an alpha check? Or is this some kind of awkward attempt at asking for my number?",
		", were you wanting a radio check, an alpha check, or the dinner check? I'm not sure what you're asking for.",
		", were you wanting a radio check, an alpha check, or a sanity check?",
		", what kind of check did you want? A radio check, an alpha check, or a coat check?",
		", what kind of check did you want? A radio check, an alpha check, or a reality check?",
		", I can't tell if you wanted a radio check or an alpha check.",
	}

	reply := c.ComposeCallsigns(response.Callsign) + replies[rand.IntN(len(replies))]
	return NaturalLanguageResponse{
		Subtitle: reply,
		Speech:   reply,
	}
}
