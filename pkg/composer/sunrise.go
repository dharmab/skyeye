package composer

import (
	"fmt"
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeSunriseCall implements [Composer.ComposeSunriseCall].
func (c *composer) ComposeSunriseCall(call brevity.SunriseCall) NaturalLanguageResponse {
	message := NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("All players: GCI %s (bot) sunrise on ", c.callsign),
		Speech:   fmt.Sprintf("All players, GCI %s sunrise on ", c.callsign),
	}

	writeBoth := func(s string) {
		message.Subtitle += s
		message.Speech += s
	}

	for i := range len(call.Frequencies) {
		frequency := call.Frequencies[i]
		decimal := fmt.Sprintf("%.3f", frequency.Megahertz())
		decimal = strings.TrimRight(decimal, "0")
		if strings.HasSuffix(decimal, ".") {
			decimal += "0"
		}
		message.Subtitle += decimal
		splits := strings.Split(decimal, ".")
		message.Speech += PronounceDecimal(frequency.Megahertz(), len(splits[1]), "point")
		if len(call.Frequencies) > 1 {
			if i == len(call.Frequencies)-2 {
				writeBoth(" and ")
			} else if i < len(call.Frequencies)-2 {
				writeBoth(", ")
			}
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
