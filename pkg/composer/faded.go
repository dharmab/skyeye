package composer

import (
	"fmt"
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func (c *composer) ComposeFadedCall(call brevity.FadedCall) NaturalLanguageResponse {
	var subtitleBuilder strings.Builder
	var speechBuilder strings.Builder

	writeBoth := func(s string) {
		subtitleBuilder.WriteString(s)
		speechBuilder.WriteString(s)
	}

	g := call.Group()

	if g.Contacts() == 1 {
		writeBoth("single contact")
	} else {
		writeBoth(fmt.Sprintf("%d contacts", g.Contacts()))
	}
	writeBoth(" faded bullseye")
	bullseye := c.ComposeBullseye(g.Bullseye())
	subtitleBuilder.WriteString(fmt.Sprintf(" %s", bullseye.Subtitle))
	speechBuilder.WriteString(fmt.Sprintf(" %s", bullseye.Speech))

	if g.Track() != brevity.UnknownDirection {
		writeBoth(fmt.Sprintf(", track %s", g.Track()))
	}

	if g.Type() != "" {
		writeBoth(fmt.Sprintf(", %s", g.Type()))
	}

	subtitleBuilder.WriteString(".")

	return NaturalLanguageResponse{
		Subtitle: subtitleBuilder.String(),
		Speech:   speechBuilder.String(),
	}
}
