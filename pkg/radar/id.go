package radar

import (
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/trackfiles"
)

func (s *scope) FindCallsign(callsign string, coalition coalitions.Coalition) (string, *trackfiles.Trackfile) {
	foundCallsign, tf, ok := s.contacts.getByCallsignAndCoalititon(callsign, coalition)
	if !ok {
		return callsign, nil
	}
	return foundCallsign, tf
}

func (s *scope) FindUnit(unitId uint32) *trackfiles.Trackfile {
	trackfile, ok := s.contacts.getByUnitID(unitId)
	if !ok {
		return nil
	}
	return trackfile
}
