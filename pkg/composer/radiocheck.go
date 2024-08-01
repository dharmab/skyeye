package composer

import (
	"fmt"
	"math/rand"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeRadioCheckResponse implements [Composer.ComposeRadioCheckResponse].
func (c *composer) ComposeRadioCheckResponse(response brevity.RadioCheckResponse) NaturalLanguageResponse {
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

	variation := replies[rand.Intn(len(replies))]
	reply := fmt.Sprintf(variation, response.Callsign)
	return NaturalLanguageResponse{
		Subtitle: reply,
		Speech:   reply,
	}
}
