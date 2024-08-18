package composer

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
)

// ComposeCoreInformationFormat communicates information about groups.
// Reference: ATP 3-52.4 chapter IV section 3
func (c *composer) ComposeCoreInformationFormat(groups ...brevity.Group) NaturalLanguageResponse {
	if len(groups) == 0 {
		return NaturalLanguageResponse{
			Subtitle: string(brevity.Clean),
			Speech:   string(brevity.Clean),
		}
	}

	response := NaturalLanguageResponse{}
	for _, group := range groups {
		groupResponse := c.ComposeGroup(group)
		response.Speech += groupResponse.Speech
		response.Subtitle += groupResponse.Subtitle
	}

	return response
}

func (c *composer) ComposeGroup(group brevity.Group) NaturalLanguageResponse {
	if group.BRAA() != nil && !group.BRAA().Bearing().IsMagnetic() {
		log.Error().Stringer("bearing", group.BRAA().Bearing()).Msg("bearing provided to ComposeGroup should be magnetic")
	}
	if group.Bullseye() != nil && !group.Bullseye().Bearing().IsMagnetic() {
		log.Error().Stringer("bearing", group.Bullseye().Bearing()).Msg("bearing provided to ComposeGroup should be magnetic")
	}
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
	stacks := group.Stacks()
	if bullseye := group.Bullseye(); bullseye != nil {
		bullseye := c.ComposeBullseye(*bullseye)
		altitude := c.ComposeAltitudeStacks(stacks, group.Declaration())
		speech.WriteString(fmt.Sprintf("%s %s, %s", label, bullseye.Speech, altitude))
		subtitle.WriteString(fmt.Sprintf("%s %s, %s", label, bullseye.Subtitle, altitude))
		if group.Track() != brevity.UnknownDirection {
			writeBoth(fmt.Sprintf(", track %s", group.Track()))
		}
	} else if group.BRAA() != nil {
		braa := c.ComposeBRAA(group.BRAA(), group.Declaration())
		speech.WriteString(fmt.Sprintf("%s %s", label, braa.Speech))
		subtitle.WriteString(fmt.Sprintf("%s %s", label, braa.Subtitle))
		isCardinalAspect := slices.Contains([]brevity.Aspect{brevity.Flank, brevity.Beam, brevity.Drag}, group.BRAA().Aspect())
		isTrackKnown := group.Track() != brevity.UnknownDirection
		if isCardinalAspect && isTrackKnown {
			writeBoth(fmt.Sprintf(" %s", group.Track()))
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

	if !group.High() {
		if len(stacks) > 1 {
			writeBoth(fmt.Sprintf(", %s", c.ComposeAltitudeFillIns(stacks)))
		}
	}

	// Platform
	if len(group.Platforms()) > 0 {
		writeBoth(", ")
		writeBoth(strings.Join(group.Platforms(), ", "))
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

	writeBoth(". ")

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

func (c *composer) ComposeAltitudeStacks(stacks []brevity.Stack, declaration brevity.Declaration) string {
	if len(stacks) == 0 {
		return "altitude unknown"
	}

	if len(stacks) == 1 {
		return c.ComposeAltitude(stacks[0].Altitude, declaration)
	}

	s := fmt.Sprintf("stack %s", c.ComposeAltitude(stacks[0].Altitude, declaration))
	for i := 1; i < len(stacks)-1; i++ {
		s += fmt.Sprintf(", %s", c.ComposeAltitude(stacks[i].Altitude, declaration))
	}
	s += fmt.Sprintf(", and %s", c.ComposeAltitude(stacks[len(stacks)-1].Altitude, declaration))
	return s
}

func (c *composer) ComposeAltitudeFillIns(stacks []brevity.Stack) string {
	if len(stacks) == 2 {
		return fmt.Sprintf("%d high, %d low", stacks[0].Count, stacks[1].Count)
	}

	if len(stacks) == 3 {
		return fmt.Sprintf("%d high, %d medium, %d low", stacks[0].Count, stacks[1].Count, stacks[2].Count)
	}
	return ""
}

func (c *composer) ComposeAltitude(altitude unit.Length, declaration brevity.Declaration) string {
	if altitude.Meters() < 100 {
		return "altitude unknown"
	}
	if declaration == brevity.Friendly {
		if altitude.Feet() < 1000 {
			hundreds := int(math.Round(altitude.Feet() / 100))
			return fmt.Sprintf("cherubs %d", hundreds)
		}
		thousands := int(math.Round(altitude.Feet() / 1000))
		return fmt.Sprintf("angels %d", thousands)
	}
	if altitude > 1000 {
		return fmt.Sprint(int(math.Round(altitude.Feet()/1000) * 1000))
	}
	return fmt.Sprint(int(math.Round(altitude.Feet()/100) * 100))
}
