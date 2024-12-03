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
	for i, group := range groups {
		groupResponse := c.ComposeGroup(group)
		response.Speech += groupResponse.Speech
		response.Subtitle += groupResponse.Subtitle
		if i < len(groups)-1 {
			response.Speech += " "
			response.Subtitle += " "
		}
	}

	return response
}

func (c *composer) ComposeGroup(group brevity.Group) (response NaturalLanguageResponse) {
	if group.BRAA() != nil && !group.BRAA().Bearing().IsMagnetic() {
		log.Error().Stringer("bearing", group.BRAA().Bearing()).Msg("bearing provided to ComposeGroup should be magnetic")
	}
	if group.Bullseye() != nil && !group.Bullseye().Bearing().IsMagnetic() {
		log.Error().Stringer("bearing", group.Bullseye().Bearing()).Msg("bearing provided to ComposeGroup should be magnetic")
	}
	label := "Group"
	if group.Threat() {
		label = "Group threat"
	}

	// Group location, altitude, and track direction or specific aspect
	stacks := group.Stacks()
	isTrackKnown := group.Track() != brevity.UnknownDirection
	if group.Bullseye() != nil {
		bullseye := c.ComposeBullseye(*group.Bullseye())
		altitude := c.ComposeAltitudeStacks(stacks, group.Declaration())
		response.Write(
			fmt.Sprintf("%s %s, %s", label, bullseye.Speech, altitude),
			fmt.Sprintf("%s %s, %s", label, bullseye.Subtitle, altitude),
		)
		if isTrackKnown {
			response.WriteBoth(fmt.Sprintf(", track %s", group.Track()))
		}
	} else if group.BRAA() != nil {
		braa := c.ComposeBRAA(group.BRAA(), group.Declaration())
		response.Write(
			fmt.Sprintf("%s %s", label, braa.Speech),
			fmt.Sprintf("%s %s", label, braa.Subtitle),
		)
		isCardinalAspect := slices.Contains([]brevity.Aspect{brevity.Flank, brevity.Beam, brevity.Drag}, group.BRAA().Aspect())
		if isCardinalAspect && isTrackKnown {
			response.WriteBoth(fmt.Sprintf(" %s", group.Track()))
		}
	}

	// Declaration
	response.WriteBoth(fmt.Sprintf(", %s", group.Declaration()))
	if group.MergedWith() == 1 {
		response.WriteBoth(", merged with 1 friendly")
	}
	if group.MergedWith() > 1 {
		response.WriteBoth(fmt.Sprintf(", merged with %d friendlies", group.MergedWith()))
	}

	// Fill-in information

	// Heavy and number of contacts
	if group.Heavy() {
		response.WriteBoth(", heavy")
	}
	contacts := c.ComposeContacts(group.Contacts())
	response.WriteResponse(contacts)

	if !group.High() {
		if len(stacks) > 1 {
			response.WriteBoth(", " + c.ComposeAltitudeFillIns(stacks))
		}
	}

	// Platform
	if len(group.Platforms()) > 0 {
		response.WriteBoth(", ")
		response.WriteBoth(strings.Join(group.Platforms(), ", "))
	}

	// High
	if group.High() {
		response.WriteBoth(", high")
	}

	// Fast or very fast
	if group.Fast() {
		response.WriteBoth(", fast")
	} else if group.VeryFast() {
		response.WriteBoth(", very fast")
	}

	response.WriteBoth(".")
	return
}

// ComposeContacts communicates the number of contacts in a group.
// Reference: ATP 3-52.4 chapter IV section 2.
func (_ *composer) ComposeContacts(n int) NaturalLanguageResponse {
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
	s += " and " + c.ComposeAltitude(stacks[len(stacks)-1].Altitude, declaration)
	return s
}

func (_ *composer) ComposeAltitudeFillIns(stacks []brevity.Stack) string {
	if len(stacks) == 2 {
		return fmt.Sprintf("%d high, %d low", stacks[0].Count, stacks[1].Count)
	}

	if len(stacks) == 3 {
		return fmt.Sprintf("%d high, %d medium, %d low", stacks[0].Count, stacks[1].Count, stacks[2].Count)
	}
	return ""
}

func (_ *composer) ComposeAltitude(altitude unit.Length, declaration brevity.Declaration) string {
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
