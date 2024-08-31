package composer

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
)

// ComposeCoreInformationFormat communicates information about groups.
// Reference: ATP 3-52.4 chapter IV section 3.
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
		isFurball := group.Declaration() == brevity.Furball
		if isCardinalAspect && isTrackKnown && !isFurball {
			writeBoth(fmt.Sprintf(" %s", group.Track()))
		}
	}

	// Declaration
	writeBoth(fmt.Sprintf(", %s", group.Declaration()))
	if group.MergedWith() == 1 {
		writeBoth(", merged with 1 friendly")
	}
	if group.MergedWith() > 1 {
		writeBoth(fmt.Sprintf(", merged with %d friendlies", group.MergedWith()))
	}

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
			writeBoth(", " + c.ComposeAltitudeFillIns(stacks))
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

// ComposeMergedWithGroup is a short form of describing a group for use in merge calls.
func (c *composer) ComposeMergedWithGroup(group brevity.Group) NaturalLanguageResponse {
	var speech, subtitle strings.Builder
	if group.Contacts() > 1 {
		contacts := c.ComposeContacts(group.Contacts())
		speech.WriteString(contacts.Speech + ", ")
		subtitle.WriteString(contacts.Subtitle + ", ")
	}

	if group.MergedWith() > 0 {
		mergedWith := " merged with 1 other friendly"
		if group.MergedWith() > 1 {
			mergedWith = fmt.Sprintf(" merged with %d other friendlies", group.MergedWith())
		}
		speech.WriteString(mergedWith)
		subtitle.WriteString(mergedWith)
	}

	return NaturalLanguageResponse{
		Subtitle: subtitle.String(),
		Speech:   speech.String(),
	}
}

// ComposeContacts communicates the number of contacts in a group.
// Reference: ATP 3-52.4 chapter IV section 2.
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

	s := "stack " + c.ComposeAltitude(stacks[0].Altitude, declaration)
	for i := 1; i < len(stacks)-1; i++ {
		s += ", " + c.ComposeAltitude(stacks[i].Altitude, declaration)
	}
	s += ", and " + c.ComposeAltitude(stacks[len(stacks)-1].Altitude, declaration)
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
	hundreds := int(math.Round(altitude.Feet() / 100))
	thousands := int(math.Round(altitude.Feet() / 1000))
	if hundreds == 0 {
		return "altitude unknown"
	}

	if declaration == brevity.Friendly {
		if altitude < 1000*unit.Foot {
			return fmt.Sprintf("cherubs %d", hundreds)
		}
		return fmt.Sprintf("angels %d", thousands)
	}

	if altitude < 1000*unit.Foot {
		return strconv.Itoa(hundreds * 100)
	}
	return strconv.Itoa(thousands * 1000)
}
