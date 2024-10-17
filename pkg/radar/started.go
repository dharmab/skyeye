package radar

import "github.com/rs/zerolog/log"

func (s *scope) handleStarted() {
	log.Info().Msg("clearing all trackfiles due to mission (re)start")
	s.contacts.reset()
	s.callbackLock.RLock()
	defer s.callbackLock.RUnlock()
	if s.startedCallback != nil {
		s.startedCallback()
	}
}
