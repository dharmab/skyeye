package composer

import (
	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeFadedCall constructs natural language brevity for announcing a contact has faded.
func (c *Composer) ComposeFadedCall(call brevity.FadedCall) (response NaturalLanguageResponse) {
	response.WriteBoth(c.composeCallsigns(c.Callsign) + ", ")
	if call.Group.Contacts() == 1 {
		response.WriteBoth("single contact faded,")
	} else {
		response.WriteBothf("%d contacts faded,", call.Group.Contacts())
	}

	if bullseye := call.Group.Bullseye(); bullseye != nil {
		bullseye := c.composeBullseye(bullseye)
		response.WriteResponse(bullseye)
	}

	if call.Group.Track() != brevity.UnknownDirection {
		response.WriteBothf(", track %s", call.Group.Track())
	}

	if call.Group.Declaration() != brevity.Unable {
		response.WriteBothf(", %s", call.Group.Declaration())
	}

	for _, platform := range call.Group.Platforms() {
		response.WriteBoth(", " + platform)
	}

	response.WriteBoth(".")
	return
}
