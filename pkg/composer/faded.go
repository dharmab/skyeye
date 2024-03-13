package composer

import (
	"fmt"
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func (c *composer) ComposeFadedCall(call brevity.FadedCall) NaturalLanguageResponse {
	var subtitle, speech strings.Builder

	writeBoth := func(s string) {
		subtitle.WriteString(s)
		speech.WriteString(s)
	}

	group := call.Group()

	if group.Contacts() == 1 {
		writeBoth("single contact")
	} else {
		writeBoth(fmt.Sprintf("%d contacts", group.Contacts()))
	}
	writeBoth(" faded bullseye")
	bullseye := c.ComposeBullseye(group.Bullseye())
	subtitle.WriteString(fmt.Sprintf(" %s", bullseye.Subtitle))
	speech.WriteString(fmt.Sprintf(" %s", bullseye.Speech))

	if group.Track() != brevity.UnknownDirection {
		writeBoth(fmt.Sprintf(", track %s", group.Track()))
	}

	if group.Type() != "" {
		writeBoth(fmt.Sprintf(", %s", group.Type()))
	}

	subtitle.WriteString(".")

	return NaturalLanguageResponse{
		Subtitle: subtitle.String(),
		Speech:   speech.String(),
	}
}
