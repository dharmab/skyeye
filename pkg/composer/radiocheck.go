package composer

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeRadioCheckResponse implements [Composer.ComposeRadioCheckResponse].
func (c *composer) ComposeRadioCheckResponse(response brevity.RadioCheckResponse) NaturalLanguageResponse {
	var reply string
	if response.RadarContact {
		replies := []string{
			"%s, 5 by 5.",
			"%s, 5 by 5!",
			"%s, I read you 5 by 5.",
			"%s, I've got you 5 by 5.",
			"%s, loud and clear.",
			"%s, I read you loud and clear.",
			"%s, I've got you loud and clear.",
			"%s, Lima Charlie.",
			"%s, Lima Charlie!",
		}
		reply = replies[rand.IntN(len(replies))]
	} else {
		replies1 := []string{
			"%s, I've got you 5 by 5",
			"%s, I read you 5 by 5",
			"%s, I've got you loud and clear",
			"%s, I read you loud and clear",
			"%s, I heard you",
		}
		replies2 := []string{
			"but I don't see you on the scope.",
			"but I don't see you on the radar.",
			"but I don't see you on the scope.",
			"but I don't see you on the radar.",
			"but you are not on the scope.",
			"but you are not on my radar.",
		}
		reply = fmt.Sprintf("%s, %s", replies1[rand.IntN(len(replies1))], replies2[rand.IntN(len(replies2))])
	}
	reply = fmt.Sprintf(reply, strings.ToUpper(response.Callsign))
	return NaturalLanguageResponse{
		Subtitle: reply,
		Speech:   reply,
	}
}
