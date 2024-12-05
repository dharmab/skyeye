package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeFadedCall constructs natural language brevity for announcing a contact has faded.
func (c *Composer) ComposeFadedCall(call brevity.FadedCall) (response NaturalLanguageResponse) {
	response.WriteBoth(c.composeCallsigns(c.Callsign) + ", ")
	if call.Group.Contacts() == 1 {
		response.WriteBoth("single contact faded,")
	} else {
		response.WriteBoth(fmt.Sprintf("%d contacts faded,", call.Group.Contacts()))
	}

	if bullseye := call.Group.Bullseye(); bullseye != nil {
		bullseye := c.ComposeBullseye(*bullseye)
		response.WriteResponse(bullseye)
	}

	if call.Group.Track() != brevity.UnknownDirection {
		response.WriteBoth(fmt.Sprintf(", track %s", call.Group.Track()))
	}

	if call.Group.Declaration() != brevity.Unable {
		response.WriteBoth(fmt.Sprintf(", %s", call.Group.Declaration()))
	}

	for _, platform := range call.Group.Platforms() {
		response.WriteBoth(", " + platform)
	}

	response.WriteBoth(".")
	return
}
