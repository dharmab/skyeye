package radar

import (
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/rs/zerolog/log"
)

func (s *scope) FindCallsign(callsign string, coalition coalitions.Coalition) (string, *trackfiles.Trackfile) {
	log.Debug().Str("callsign", callsign).Any("contacts", s.contacts).Msg("searching scope for trackfile matching callsign")
	foundCallsign, tf, ok := s.contacts.getByCallsignAndCoalititon(callsign, coalition)
	if !ok {
		return callsign, nil
	}
	return foundCallsign, tf
}

func (s *scope) FindUnit(unitId uint32) *trackfiles.Trackfile {
	log.Debug().Uint32("unitId", unitId).Any("contacts", s.contacts).Msg("searching scope for trackfile matching unitId and coalition")
	trackfile, ok := s.contacts.getByUnitID(unitId)
	if !ok {
		return nil
	}
	return trackfile
}
