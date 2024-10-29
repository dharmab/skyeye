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

	writeBoth(strings.ToUpper(c.callsign) + ", ")
	if call.Group.Contacts() == 1 {
		writeBoth("single contact faded,")
	} else {
		writeBoth(fmt.Sprintf("%d contacts faded,", call.Group.Contacts()))
	}

	if bullseye := call.Group.Bullseye(); bullseye != nil {
		bullseye := c.ComposeBullseye(*bullseye)
		subtitle.WriteString(" " + bullseye.Subtitle)
		speech.WriteString(" " + bullseye.Speech)
	}

	if call.Group.Track() != brevity.UnknownDirection {
		writeBoth(fmt.Sprintf(", track %s", call.Group.Track()))
	}

	if call.Group.Declaration() != brevity.Unable {
		writeBoth(fmt.Sprintf(", %s", call.Group.Declaration()))
	}

	for _, platform := range call.Group.Platforms() {
		writeBoth(", " + platform)
	}

	writeBoth(".")

	return NaturalLanguageResponse{
		Subtitle: subtitle.String(),
		Speech:   speech.String(),
	}
}
