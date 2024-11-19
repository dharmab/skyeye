package composer

import (
	"fmt"
	"math/rand/v2"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func (c *composer) ComposeShoppingResponse(r brevity.ShoppingResponse) NaturalLanguageResponse {
	replies1 := []string{
		"%s, SHOPPING is a brevity code related to air-to-ground operations. It is not an air-to-air combat term.",
		"%s, I'm an air battle manager, not a JTAC or Forward Air Controller. SHOPPING is an air-to-surface brevity code.",
		"%s, if you flip to page 38 of MULTI-SERVICE TACTICS, TECHNIQUES, AND PROCEDURES FOR MULTI-SERVICE BREVITY CODES, you'll see SHOPPING is an air-to-surface brevity code used by JTACs and FACs.",
		"%s, SHOPPING is something you'd request from a JTAC or FAC. It's specific to air-to-surface warfare.",
		"%s, the brevity code SHOPPING means you're requesting a ground target. As an air battle manager, I'm not the right person to ask for that.",
		"%s, why am I not surprised that you didn't read the manual? SHOPPING is a brevity code for air-to-surface warfare, not air-to-air combat.",
	}
	variation1 := replies1[rand.IntN(len(replies1))]

	replies2 := []string{
		"You can ask for a BOGEY DOPE for information on the nearest air threat, or a PICTURE for a ranked list of the most immediate threats to your coalition.",
		"Instead, you can ask me for a BOGEY DOPE to get information on the nearest air threat, or a PICTURE for a ranked list of the most immediate threats to your coalition.",
		"What I can help you with is a BOGEY DOPE for information on the nearest air threat, or a PICTURE for a ranked list of the most immediate threats to your coalition.",
		"Instead of SHOPPING, you can ask for a BOGEY DOPE for information on the nearest air threat, or a PICTURE for a ranking of the most immediate threats to your coalition.",
	}
	variation2 := replies2[rand.IntN(len(replies2))]

	reply := fmt.Sprintf(
		fmt.Sprintf("%s %s", variation1, variation2),
		c.ComposeCallsigns(r.Callsign),
	)
	return NaturalLanguageResponse{
		Subtitle: reply,
		Speech:   reply,
	}
}
