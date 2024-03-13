package composer

import (
	"fmt"
	"math/rand"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func (c *composer) ComposeRadioCheckResponse(r brevity.RadioCheckResponse) NaturalLanguageResponse {
	var replies []string
	if r.Status() {
		replies = []string{
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
	} else {
		replies = []string{
			"%q...? Say again.",
			"%s, I didn't catch that.",
			"%s, unable to read you. Say again.",
			"%q...? Please repeat.",
			"%s, negative, say again.",
			"%s, negative, repeat last.",
			"%s, signal is weak. Say again.",
			"%s, poor signal. Say again.",
		}
	}

	f := replies[rand.Intn(len(replies))]
	s := fmt.Sprintf(f, r.Callsign())
	return NaturalLanguageResponse{
		Subtitle: s,
		Speech:   s,
	}
}
