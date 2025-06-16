package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
)

func (c *Composer) composeCorrelation(requestType, callsign string, status bool, bearing bearings.Bearing, grp brevity.Group) NaturalLanguageResponse {
	if status {
		nlr := NaturalLanguageResponse{}

		callsigns := c.composeCallsigns(callsign)
		nlr.WriteBoth(callsigns)

		_range := int(grp.BRAA().Range().NauticalMiles())
		nlr.WriteBothf(", %s range %d", requestType, _range)

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
	if bearing == nil {
		nlr := NaturalLanguageResponse{}
		message := fmt.Sprintf("%s, %s", c.composeCallsigns(callsign), brevity.Unable)
		nlr.WriteBoth(message)
		return nlr
	}
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("%s, %s clean %d.", c.composeCallsigns(callsign), c.composeCallsigns(c.Callsign), int(bearing.Degrees())),
		Speech:   fmt.Sprintf("%s, %s, clean %s", c.composeCallsigns(callsign), c.composeCallsigns(c.Callsign), pronounceBearing(bearing)),
	}
}
