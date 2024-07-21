package composer

import (
	"fmt"
	"math/rand"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func (c *composer) ComposeRadioCheckResponse(r brevity.RadioCheckResponse) NaturalLanguageResponse {

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

	f := replies[rand.Intn(len(replies))]
	s := fmt.Sprintf(f, r.Callsign)
	return NaturalLanguageResponse{
		Subtitle: s,
		Speech:   s,
	}
}
