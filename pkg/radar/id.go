package radar

import (
	"time"

	"github.com/dharmab/skyeye/pkg/trackfile"
	"github.com/rs/zerolog/log"
)

func (s *scope) FindCallsign(callsign string) *trackfile.Trackfile {
	s.lock.Lock()
	defer s.lock.Unlock()
	logger := log.With().Str("callsign", callsign).Logger()
	logger.Debug().Any("contacts", s.contacts).Msg("searching scope for trackfile matching callsign")
	unitID, ok := s.callsignIdx[callsign]
	if !ok {
		logger.Debug().Msg("callsign not found in index")
		return nil
	}
	logger = logger.With().Int("unitID", int(unitID)).Logger()
	tf, ok := s.contacts[unitID]
	if !ok {
		logger.Debug().Msg("unitID not found in contacts")
		return nil
	}
	if tf.LastKnown().Timestamp.Before(time.Now().Add(-1 * time.Minute)) {
		logger.Debug().Msg("trackfile is stale")
		return nil
	}
	logger.Debug().Msg("found trackfile")
	return tf
}

func (s *scope) FindUnit(unitId uint32) *trackfile.Trackfile {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, tf := range s.contacts {
		if tf.Contact.UnitID == unitId {
			return tf
		}
	}
	return nil
}
