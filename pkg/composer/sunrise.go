package composer

import (
	"fmt"
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeSunriseCall constructs natural language brevity for announcing GCI services are online.
func (c *Composer) ComposeSunriseCall(call brevity.SunriseCall) NaturalLanguageResponse {
	message := NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("All players: GCI %s (bot) sunrise on ", c.composeCallsigns(c.Callsign)),
		Speech:   fmt.Sprintf("All players, GCI %s sunrise on ", c.composeCallsigns(c.Callsign)),
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
		message.Speech += pronounceDecimal(frequency.Megahertz(), len(splits[1]), "point")
		if len(call.Frequencies) > 1 {
			if i == len(call.Frequencies)-2 {
				message.WriteBoth(" and ")
			} else if i < len(call.Frequencies)-2 {
				message.WriteBoth(", ")
			}
		}
	}

	return message
}
