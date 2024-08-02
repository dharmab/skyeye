package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeSunriseCall implements [Composer.ComposeSunriseCall].
func (c *composer) ComposeSunriseCall(call brevity.SunriseCall) NaturalLanguageResponse {
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("All players: GCI %s (bot) sunrise on %.3fMHz", c.callsign, call.Frequency.Megahertz()),
		Speech:   fmt.Sprintf("All players, GCI %s sunrise on %s", c.callsign, PronounceDecimal(call.Frequency.Megahertz(), 3, "point")),
	}
}

func (c *composer) ComposeMidnightCall(call brevity.MidnightCall) NaturalLanguageResponse {
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("All players: GCI %s midnight. See ya!", c.callsign),
		Speech:   fmt.Sprintf("All players, GCI %s midnight. sssssssseeeeya!", c.callsign),
	}
}
