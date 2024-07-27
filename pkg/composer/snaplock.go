package composer

import (
	"fmt"
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeSnaplockResponse implements [Composer.ComposeSnaplockResponse].
func (c *composer) ComposeSnaplockResponse(r brevity.SnaplockResponse) NaturalLanguageResponse {
	isLocationMissing := r.Group.BRAA() == nil
	isDeclarationUnable := r.Group.Declaration() == brevity.Unable
	isDeclarationFurball := r.Group.Declaration() == brevity.Furball
	if isLocationMissing || isDeclarationUnable || isDeclarationFurball {
		reply := fmt.Sprintf("%s, %s", r.Callsign, brevity.Unable)
		return NaturalLanguageResponse{
			Subtitle: reply,
			Speech:   reply,
		}
	}

	if r.Status {
		var subtitleBuilder strings.Builder
		var speechBuilder strings.Builder
		braa := c.ComposeBRAA(r.Group.BRAA())
		contacts := c.ComposeContacts(r.Group.Contacts())
		subtitleBuilder.WriteString(fmt.Sprintf(
			"%s, %s, %s",
			r.Callsign,
			braa.Subtitle,
			r.Group.Declaration(),
		))
		speechBuilder.WriteString(fmt.Sprintf(
			"%s, %s, %s",
			r.Callsign,
			braa.Speech,
			r.Group.Declaration(),
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

	reply := fmt.Sprintf("%s, %s", r.Callsign, brevity.Clean)
	return NaturalLanguageResponse{
		Subtitle: reply,
		Speech:   reply,
	}
}
