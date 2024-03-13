package composer

import (
	"fmt"
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeCoreInformationFormat communicates information about groups.
// Reference: ATP 3-52.4 chapter IV section 3
func (c *composer) ComposeCoreInformationFormat(groups []brevity.Group) NaturalLanguageResponse {
	var speech, subtitle strings.Builder

	writeBoth := func(s string) {
		speech.WriteString(s)
		subtitle.WriteString(s)
	}

	// Total number of groups
	if len(groups) > 1 {
		writeBoth(fmt.Sprintf("%d groups. ", len(groups)))
	}

	for i, group := range groups {
		if i > 0 {
			writeBoth(" ")
		}
		groupResponse := c.ComposeGroup(group)
		speech.WriteString(groupResponse.Speech)
		subtitle.WriteString(groupResponse.Subtitle)
		writeBoth(".")
	}

	return NaturalLanguageResponse{
		Subtitle: subtitle.String(),
		Speech:   speech.String(),
	}
}

func (c *composer) ComposeGroup(group brevity.Group) NaturalLanguageResponse {
	var speech, subtitle strings.Builder

	writeBoth := func(s string) {
		speech.WriteString(s)
		subtitle.WriteString(s)
	}

	label := "Group"
	if group.Threat() {
		label = "Group threat"
	}

	// Group location, altitude, and track direction or specific aspect
	if group.Bullseye() != nil {
		bullseye := c.ComposeBullseye(group.Bullseye())

		speech.WriteString(fmt.Sprintf("%s %s, %d", label, bullseye.Speech, int(group.Altitude().Feet())))
		subtitle.WriteString(fmt.Sprintf("%s %s/%d", label, bullseye.Subtitle, int(group.Altitude().Feet())))
		if group.Track() != brevity.UnknownDirection {
			writeBoth(fmt.Sprintf(", track %s", group.Track()))
		}
	} else if group.BRAA() != nil {
		braa := c.ComposeBRAA(group.BRAA())
		speech.WriteString(fmt.Sprintf("%s %s, %d", label, braa.Speech, int(group.Altitude().Feet())))
		subtitle.WriteString(fmt.Sprintf("%s %s, %d", label, braa.Subtitle, int(group.Altitude().Feet())))
		if group.Aspect() != brevity.UnknownAspect {
			writeBoth(fmt.Sprintf(", %s", group.Aspect()))
		}
	}

	// Declaration
	writeBoth(fmt.Sprintf(", %s", group.Declaration()))

	// Fill-in information

	// Heavy and number of contacts
	if group.Heavy() {
		writeBoth(", heavy")
	}
	contacts := c.ComposeContacts(group.Contacts())
	subtitle.WriteString(contacts.Subtitle)
	speech.WriteString(contacts.Speech)

	// Platform/type
	if group.Type() != "" {
		writeBoth(fmt.Sprintf(", %s", group.Type()))
	}

	// High
	if group.High() {
		writeBoth(", high")
	}

	// Fast or very fast
	if group.Fast() {
		writeBoth(", fast")
	} else if group.VeryFast() {
		writeBoth(", very fast")
	}

	return NaturalLanguageResponse{
		Subtitle: subtitle.String(),
		Speech:   speech.String(),
	}
}

// ComposeContacts communicates the number of contacts in a group.
// Reference: ATP 3-52.4 chapter IV section 2
func (c *composer) ComposeContacts(n int) NaturalLanguageResponse {
	// single contact is assumed if unspecified
	s := ""
	if n > 1 {
		s = fmt.Sprintf(", %d contacts", n)
	}
	return NaturalLanguageResponse{
		Subtitle: s,
		Speech:   s,
	}
}
