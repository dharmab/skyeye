package composer

import (
	"fmt"
	"math/rand"

	"github.com/dharmab/skyeye/pkg/brevity"
)

var radioCheckOKReplies = []string{
	"%s, 5 by 5.",
	"%s, I read you 5 by 5.",
	"%s, loud and clear.",
	"%s, I read you loud and clear.",
	"%s, I've got you loud and clear.",
	"%s, Lima Charlie.",
	"%s, copy.",
	"%s, solid copy.",
	"%s, contact.",
}

var radioCheckFailReplies = []string{
	"%q...? Say again.",
	"%s, I didn't catch that.",
	"%s, unable to read you. Say again.",
	"%q...? Please repeat.",
	"%s, negative, say again.",
	"%s, negative, repeat last.",
	"%s, signal is weak. Say again.",
	"%s, poor signal. Say again.",
}

func (c *composer) ComposeRadioCheckResponse(r brevity.RadioCheckResponse) NaturalLanguageResponse {
	var f string
	if r.Status() {
		// pick a random OK reply
		f = radioCheckOKReplies[rand.Intn(len(radioCheckOKReplies))]
	} else {
		// pick a random fail reply
		f = radioCheckFailReplies[rand.Intn(len(radioCheckFailReplies))]

	}
	s := fmt.Sprintf(f, r.Callsign())
	return NaturalLanguageResponse{
		Subtitle: s,
		Speech:   s,
	}
}
