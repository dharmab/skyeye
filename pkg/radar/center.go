package radar

import (
	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/encyclopedia"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/rs/zerolog/log"
)

func shiftPointTowards(a orb.Point, b orb.Point) orb.Point {
	if a.Lon() == 0 && a.Lat() == 0 {
		return b
	}
	return geo.Midpoint(a, b)
}

func (s *scope) updateCenterPoint() {
	blue := orb.Point{}
	red := orb.Point{}
	itr := s.contacts.itr()
	for itr.next() {
		contact := itr.value()
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
	precision := 50 * unit.NauticalMile
	distance := unit.Length(geo.Distance(s.center, newCenter)) * unit.Meter
	if distance > precision {
		bearing := bearings.NewTrueBearing(unit.Angle(geo.Bearing(s.center, newCenter)) * unit.Degree)
		s.center = newCenter
		log.Info().
			Float64("lon", s.center.Lon()).
			Float64("lat", s.center.Lat()).
			Msgf("center point shifted %.1f NM along bearing %.0f", distance.NauticalMiles(), bearing.RoundedDegrees())
	}
}
