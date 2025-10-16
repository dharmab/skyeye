package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeSpikedResponse constructs natural language brevity for responding to a SPIKED call.
func (c *Composer) ComposeSpikedResponse(response brevity.SpikedResponseV2) NaturalLanguageResponse {
	if response.Status {
		nlr := NaturalLanguageResponse{}

		callsigns := c.composeCallsigns(response.Callsign)
		nlr.WriteBoth(callsigns)

		grp := response.Group

		_range := int(grp.BRAA().Range().NauticalMiles())
		nlr.WriteBothf(", spike range %d", _range)

		nlr.WriteBoth(", ")
		altitude := c.composeAltitudeStacks(grp.Stacks(), grp.Declaration())
		nlr.WriteBoth(altitude)

		nlr.WriteBothf(", %s", grp.BRAA().Aspect())

		if grp.BRAA().Aspect().IsCardinal() && grp.Track() != brevity.UnknownDirection {
			nlr.WriteBothf(" %s", grp.Track())
		}
		declaration := c.composeDeclaration(grp)
		nlr.WriteBoth(", ")
		nlr.WriteResponse(declaration)

		fillIns := c.composeFillIns(grp)
		if len(fillIns.Subtitle) > 0 {
			nlr.WriteResponse(fillIns)
		}
		nlr.WriteBoth(".")
		return nlr
	}
	if response.Bearing == nil {
		nlr := NaturalLanguageResponse{}
		message := fmt.Sprintf("%s, %s", c.composeCallsigns(response.Callsign), brevity.Unable)
		nlr.WriteBoth(message)
		return nlr
	}
	callerCallsign := c.composeCallsigns(response.Callsign)
	controllerCallsign := c.composeCallsigns(c.Callsign)
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("%s, %s clean %d.", callerCallsign, controllerCallsign, int(response.Bearing.Degrees())),
		Speech:   fmt.Sprintf("%s, %s, clean %s", callerCallsign, controllerCallsign, pronounceBearing(response.Bearing)),
	}
}
