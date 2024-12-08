package radar

import (
	"time"

	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/rs/zerolog/log"
)

func (r *Radar) handleStarted() {
	log.Info().Msg("clearing all trackfiles due to mission (re)start")
	r.contacts.reset()

	log.Info().Msg("clearing pending FADED trackfiles due to mission (re)start")
	r.pendingFadesLock.Lock()
	defer r.pendingFadesLock.Unlock()
	r.pendingFades = make([]sim.Faded, 0)

	log.Info().Msg("clearing FADED trackfile history due to mission (re)start")
	r.completedFadesLock.Lock()
	defer r.completedFadesLock.Unlock()
	r.completedFades = make(map[uint64]time.Time)

	r.callbackLock.RLock()
	defer r.callbackLock.RUnlock()
	if r.startedCallback != nil {
		r.startedCallback()
	}
}
