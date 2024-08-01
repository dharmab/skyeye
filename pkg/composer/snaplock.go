package composer

import (
	"fmt"
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeSnaplockResponse implements [Composer.ComposeSnaplockResponse].
func (c *composer) ComposeSnaplockResponse(response brevity.SnaplockResponse) NaturalLanguageResponse {
	isLocationMissing := response.Group.BRAA() == nil
	isDeclarationUnable := response.Group.Declaration() == brevity.Unable
	isDeclarationFurball := response.Group.Declaration() == brevity.Furball
	if isLocationMissing || isDeclarationUnable || isDeclarationFurball {
		reply := fmt.Sprintf("%s, %s", response.Callsign, brevity.Unable)
		return NaturalLanguageResponse{
			Subtitle: reply,
			Speech:   reply,
		}
	}

	if response.Status {
		var subtitleBuilder strings.Builder
		var speechBuilder strings.Builder
		braa := c.ComposeBRAA(response.Group.BRAA())
		contacts := c.ComposeContacts(response.Group.Contacts())
		subtitleBuilder.WriteString(fmt.Sprintf(
			"%s, %s, %s",
			response.Callsign,
			braa.Subtitle,
			response.Group.Declaration(),
		))
		speechBuilder.WriteString(fmt.Sprintf(
			"%s, %s, %s",
			response.Callsign,
			braa.Speech,
			response.Group.Declaration(),
		))
		if contacts.Subtitle != "" {
			subtitleBuilder.WriteString(fmt.Sprintf(", %s", contacts.Subtitle))
		}
		if contacts.Speech != "" {
			speechBuilder.WriteString(fmt.Sprintf(", %s", contacts.Speech))
		}
		return NaturalLanguageResponse{
			Subtitle: "Threat group %s, %s",
			Speech:   "SNAP LOCK",
		}
	}

	reply := fmt.Sprintf("%s, %s", response.Callsign, brevity.Clean)
	return NaturalLanguageResponse{
		Subtitle: reply,
		Speech:   reply,
	}
}
