package radar

import (
	"time"

	"github.com/dharmab/skyeye/pkg/trackfile"
	"github.com/rs/zerolog/log"
)

func (s *scope) FindCallsign(callsign string) *trackfile.Trackfile {
	return find(func() (*trackfile.Trackfile, bool) {
		logger := log.With().Str("callsign", callsign).Logger()
		logger.Debug().Any("contacts", s.contacts).Msg("searching scope for trackfile matching callsign")
		tf, ok := s.contacts.getByCallsign(callsign)
		if !ok {
			return nil, false
		}
		return tf, true
	})
}

func (s *scope) FindUnit(unitId uint32) *trackfile.Trackfile {
	return find(func() (*trackfile.Trackfile, bool) {
		logger := log.With().Uint32("unitId", unitId).Logger()
		logger.Debug().Any("contacts", s.contacts).Msg("searching scope for trackfile matching unitId")
		return s.contacts.getByUnitID(unitId)
	})
}

func find(fn func() (*trackfile.Trackfile, bool)) *trackfile.Trackfile {
	tf, ok := fn()
	if !ok {
		return nil
	}
	if tf.LastKnown().Timestamp.Before(time.Now().Add(-1 * time.Minute)) {
		log.Debug().Str("name", tf.Contact.Name).Int("unitId", int(tf.Contact.UnitID)).Dur("age", time.Since(tf.LastKnown().Timestamp)).Msg("trackfile is stale")
		return nil
	}
	return tf
}
