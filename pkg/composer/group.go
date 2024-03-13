package composer

import (
	"fmt"
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeCoreInformationFormat communicates information about groups.
// Reference: ATP 3-52.4 chapter IV section 3
func (c *composer) ComposeCoreInformationFormat(groups []brevity.Group) NaturalLanguageResponse {
	var speechBuilder strings.Builder
	var subtitleBuilder strings.Builder

	writeBoth := func(s string) {
		speechBuilder.WriteString(s)
		subtitleBuilder.WriteString(s)
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
		speechBuilder.WriteString(groupResponse.Speech)
		subtitleBuilder.WriteString(groupResponse.Subtitle)
		writeBoth(".")
	}

	return NaturalLanguageResponse{
		Subtitle: subtitleBuilder.String(),
		Speech:   speechBuilder.String(),
	}
}

func (c *composer) ComposeGroup(g brevity.Group) NaturalLanguageResponse {
	var speechBuilder strings.Builder
	var subtitleBuilder strings.Builder

	writeBoth := func(s string) {
		speechBuilder.WriteString(s)
		subtitleBuilder.WriteString(s)
	}

	// Group location, altitude, and track direction or specific aspect
	if g.Bullseye() != nil {
		bullseye := c.ComposeBullseye(g.Bullseye())
		speechBuilder.WriteString(fmt.Sprintf("group %s, %d", bullseye.Speech, int(g.Altitude().Feet())))
		subtitleBuilder.WriteString(fmt.Sprintf("Group %s/%d", bullseye.Subtitle, int(g.Altitude().Feet())))
		if g.Track() != brevity.UnknownDirection {
			writeBoth(fmt.Sprintf(", track %s", g.Track()))
		}
	} else if g.BRAA() != nil {
		braa := c.ComposeBRAA(g.BRAA())
		speechBuilder.WriteString(fmt.Sprintf("group %s, %d", braa.Speech, int(g.Altitude().Feet())))
		subtitleBuilder.WriteString(fmt.Sprintf("group %s, %d", braa.Subtitle, int(g.Altitude().Feet())))
		if g.Aspect() != brevity.UnknownAspect {
			writeBoth(fmt.Sprintf(", %s", g.Aspect()))
		}
	}

	// Declaration
	writeBoth(fmt.Sprintf(", %s", g.Declaration()))

	// Fill-in information

	// Heavy and number of contacts
	if g.Heavy() {
		writeBoth(", heavy")
	}
	contacts := c.ComposeContacts(g.Contacts())
	subtitleBuilder.WriteString(contacts.Subtitle)
	speechBuilder.WriteString(contacts.Speech)

	// Platform/type
	if g.Type() != "" {
		writeBoth(fmt.Sprintf(", %s", g.Type()))
	}

	// High
	if g.High() {
		writeBoth(", high")
	}

	// Fast or very fast
	if g.Fast() {
		writeBoth(", fast")
	} else if g.VeryFast() {
		writeBoth(", very fast")
	}

	return NaturalLanguageResponse{
		Subtitle: subtitleBuilder.String(),
		Speech:   speechBuilder.String(),
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
