package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func (c *composer) ComposeSunriseCall(brevity.SunriseCall) NaturalLanguageResponse {
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("All players: GCI %s (bot) sunrise on %.3fMHz", c.callsign, c.frequency.Megahertz()),
		Speech:   fmt.Sprintf("All players, GCI %s sunrise on %s", c.callsign, PronounceDecimal(c.frequency.Megahertz(), 3, "point")),
	}
}
