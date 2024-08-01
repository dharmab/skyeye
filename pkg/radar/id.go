package radar

import (
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/rs/zerolog/log"
)

func (s *scope) FindCallsign(callsign string) (string, *trackfiles.Trackfile) {
	log.Debug().Str("callsign", callsign).Any("contacts", s.contacts).Msg("searching scope for trackfile matching callsign")
	foundCallsign, tf, ok := s.contacts.getByCallsign(callsign)
	if !ok {
		return callsign, nil
	}
	return foundCallsign, tf
}

func (s *scope) FindUnit(unitId uint32) *trackfiles.Trackfile {
	log.Debug().Uint32("unitId", unitId).Any("contacts", s.contacts).Msg("searching scope for trackfile matching unitId")
	trackfile, ok := s.contacts.getByUnitID(unitId)
	if !ok {
		return nil
	}
	return trackfile
}
