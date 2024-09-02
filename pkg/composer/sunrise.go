package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeSunriseCall implements [Composer.ComposeSunriseCall].
func (c *composer) ComposeSunriseCall(call brevity.SunriseCall) NaturalLanguageResponse {
	message := NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("All players: GCI %s (bot) sunrise on", c.callsign),
		Speech:   fmt.Sprintf("All players, GCI %s sunrise on ", c.callsign),
	}

	for i, freq := range call.Frequencies {
		message.Subtitle += fmt.Sprintf(", %f.3", freq.Megahertz())
		decimal := PronounceDecimal(freq.Megahertz(), 3, "point")
		message.Speech += fmt.Sprintf(", %s", decimal)
		if len(call.Frequencies) > 1 && i == len(call.Frequencies)-2 {
			message.Subtitle += " and"
			message.Speech += " and"
		}
	}

	return message
}

func (c *composer) ComposeMidnightCall(call brevity.MidnightCall) NaturalLanguageResponse {
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("All players: GCI %s midnight. See ya!", c.callsign),
		Speech:   fmt.Sprintf("All players, GCI %s midnight. sssssssseeeeya!", c.callsign),
	}
}
