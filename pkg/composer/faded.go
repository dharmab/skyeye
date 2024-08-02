package composer

import (
	"fmt"
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeFadedCall implements [Composer.ComposeFadedCall].
func (c *composer) ComposeFadedCall(call brevity.FadedCall) NaturalLanguageResponse {
	var subtitle, speech strings.Builder

	writeBoth := func(s string) {
		subtitle.WriteString(s)
		speech.WriteString(s)
	}

	if call.Group.Contacts() == 1 {
		writeBoth("single contact")
	} else {
		writeBoth(fmt.Sprintf("%d contacts", call.Group.Contacts()))
	}
	writeBoth(" faded")
	if bullseye := call.Group.Bullseye(); bullseye != nil {
		writeBoth(" bullseye")
		bullseye := c.ComposeBullseye(*bullseye)
		subtitle.WriteString(fmt.Sprintf(" %s", bullseye.Subtitle))
		speech.WriteString(fmt.Sprintf(" %s", bullseye.Speech))
	}

	if call.Group.Track() != brevity.UnknownDirection {
		writeBoth(fmt.Sprintf(", track %s", call.Group.Track()))
	}

	for _, platform := range call.Group.Platforms() {
		writeBoth(fmt.Sprintf(", %s", platform))
	}

	subtitle.WriteString(".")

	return NaturalLanguageResponse{
		Subtitle: subtitle.String(),
		Speech:   speech.String(),
	}
}
