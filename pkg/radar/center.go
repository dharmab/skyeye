package radar

import (
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/encyclopedia"
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/rs/zerolog/log"
)

// shiftPointTowards returns the midpoint between two points, or the second point if the first point is the origin.
func shiftPointTowards(a orb.Point, b orb.Point) orb.Point {
	if spatial.IsZero(a) {
		return b
	}
	return geo.Midpoint(a, b)
}

func (s *scope) updateCenterPoint() {
	s.centerLock.Lock()
	defer s.centerLock.Unlock()
	blue := orb.Point{}
	red := orb.Point{}
	for contact := range s.contacts.values() {
		data, ok := encyclopedia.GetAircraftData(contact.Contact.ACMIName)
		isArmed := !ok || data.ThreatRadius() > 0
		isValid := isValidTrack(contact)
		if isArmed && isValid {
			contactLocation := contact.LastKnown().Point
			switch contact.Contact.Coalition {
			case coalitions.Blue:
				blue = shiftPointTowards(blue, contactLocation)
			case coalitions.Red:
				red = shiftPointTowards(red, contactLocation)
			}
		}
	}
	var newCenter orb.Point
	isBlueOk := blue.Lon() != 0 && blue.Lat() != 0
	isRedOk := red.Lon() != 0 && red.Lat() != 0
	if isBlueOk && !isRedOk {
		newCenter = blue
	} else if !isBlueOk && isRedOk {
		newCenter = red
	} else {
		newCenter = geo.Midpoint(blue, red)
	}
	distance := spatial.Distance(s.center, newCenter)
	bearing := spatial.TrueBearing(s.center, newCenter)
	s.center = newCenter
	log.Trace().
		Float64("lon", s.center.Lon()).
		Float64("lat", s.center.Lat()).
		Msgf("center point shifted %.1f NM along bearing %.0f", distance.NauticalMiles(), bearing.RoundedDegrees())
}
