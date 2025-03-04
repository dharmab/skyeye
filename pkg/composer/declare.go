package composer

import (
	"slices"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeDeclareResponse constructs natural language brevity for responding to a DECLARE call.
func (c *Composer) ComposeDeclareResponse(response brevity.DeclareResponse) (reply NaturalLanguageResponse) {
	reply.WriteBoth(c.composeCallsigns(response.Callsign) + ", ")

	if response.Sour {
		reply.WriteBoth("unable, timber sour. Repeat your request with bullseye or BRAA position included.")
		return
	}

	if slices.Contains([]brevity.Declaration{brevity.Furball, brevity.Unable, brevity.Clean}, response.Declaration) {
		if response.Readback != nil {
			bullseye := c.composeBullseye(*response.Readback)
			reply.WriteResponse(bullseye)
			reply.WriteBoth(",")
		}
		reply.WriteBoth(" ")
		if response.Declaration == brevity.Furball {
			reply.WriteResponse(c.composeDeclaration(response.Group))
			if fillIns := c.composeFillIns(response.Group); fillIns.Subtitle != "" {
				reply.WriteResponse(fillIns)
			}
		} else {
			reply.WriteBoth(string(response.Declaration))
		}
		return
	}

	info := c.composeCoreInformationFormat(response.Group)
	reply.WriteResponse(info)

	return
}
