package radar

import "github.com/rs/zerolog/log"

func (r *Radar) handleStarted() {
	log.Info().Msg("clearing all trackfiles due to mission (re)start")
	r.contacts.reset()
	r.callbackLock.RLock()
	defer r.callbackLock.RUnlock()
	if r.startedCallback != nil {
		r.startedCallback()
	}
}
