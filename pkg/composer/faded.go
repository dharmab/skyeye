package composer

import (
	"fmt"
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeFadedCall implements [Composer.ComposeFadedCall].
func (c *composer) ComposeFadedCall(call brevity.FadedCall) NaturalLanguageResponse {
	return c.gone(call.Group, "faded")
}

// ComposeVanishedCall implements [Composer.ComposeVanishedCall].
func (c *composer) ComposeVanishedCall(call brevity.VanishedCall) NaturalLanguageResponse {
	return c.gone(call.Group, "vanished")
}

func (c *composer) gone(group brevity.Group, codeword string) NaturalLanguageResponse {
	var subtitle, speech strings.Builder

	writeBoth := func(s string) {
		subtitle.WriteString(s)
		speech.WriteString(s)
	}

	writeBoth(c.callsign + ", ")
	if group.Contacts() == 1 {
		writeBoth("single contact")
	} else {
		writeBoth(fmt.Sprintf("%d contacts", group.Contacts()))
	}
	writeBoth(" " + codeword)
	if bullseye := group.Bullseye(); bullseye != nil {
		bullseye := c.ComposeBullseye(*bullseye)
		subtitle.WriteString(fmt.Sprintf(" %s", bullseye.Subtitle))
		speech.WriteString(fmt.Sprintf(" %s", bullseye.Speech))
	}

	if group.Track() != brevity.UnknownDirection {
		writeBoth(fmt.Sprintf(", track %s", group.Track()))
	}

	if group.Declaration() != brevity.Unable {
		writeBoth(fmt.Sprintf(", %s", group.Declaration()))
	}

	for _, platform := range group.Platforms() {
		writeBoth(fmt.Sprintf(", %s", platform))
	}

	subtitle.WriteString(".")

	return NaturalLanguageResponse{
		Subtitle: subtitle.String(),
		Speech:   speech.String(),
	}
}
