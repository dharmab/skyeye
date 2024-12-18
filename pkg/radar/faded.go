package radar

import (
	"context"
	"slices"
	"time"

	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func isTrackfileInGroup(candidate *trackfiles.Trackfile, grp *group) bool {
	return slices.ContainsFunc(grp.contacts, func(member *trackfiles.Trackfile) bool {
		return member.Contact.ID == candidate.Contact.ID
	})
}

// collectFadedTrackfiles continuously collects faded contacts. When there is no new faded contact for 10 seconds,
// it collects all faded contacts into groups, removes the contacts from the database, and calls the fadedCallback.
func (r *Radar) collectFadedTrackfiles(ctx context.Context) {
	// Whenenver we pass the deadline, we collect the faded contacts into groups and call the fadedCallback.
	var deadline time.Time
	ticker := time.NewTicker(10 * time.Second)

	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case fade := <-r.fades:
			// When we receive a faded contact, we wait a little in case it's wingman is also fading.
			// This is common if the flight lands or is being engaged by a coordinated flight.
			deadline = time.Now().Add(15 * time.Second)
			go func() {
				r.pendingFadesLock.Lock()
				defer r.pendingFadesLock.Unlock()
				r.pendingFades = append(r.pendingFades, fade)
			}()
		case <-ticker.C:
			// Regularly handle pending fades.
			go func() {
				r.pendingFadesLock.Lock()
				defer r.pendingFadesLock.Unlock()
				if len(r.pendingFades) > 0 && time.Now().After(deadline) {
					log.Info().Int("count", len(r.pendingFades)).Msg("handling pending faded trackfiles")
					r.handleFaded(r.pendingFades)
					r.pendingFades = []sim.Faded{}
				}
			}()
			// Periodically clean up completed fades.
			go func() {
				r.completedFadesLock.Lock()
				defer r.completedFadesLock.Unlock()
				for id, t := range r.completedFades {
					age := time.Since(t)
					if age > 5*time.Minute {
						log.Debug().Stringer("age", age).Uint64("id", id).Msg("discarding faded trackfile from recent history")
						delete(r.completedFades, id)
					}
				}
			}()
		}
	}
}

func (r *Radar) collectFadedGroups(fades []sim.Faded) []group {
	var groups []group
	for _, fade := range fades {
		if func() bool {
			r.completedFadesLock.RLock()
			defer r.completedFadesLock.RUnlock()
			_, ok := r.completedFades[fade.ID]
			return ok
		}() {
			log.Info().Uint64("id", fade.ID).Msg("skipping faded trackfile because it was recently handled")
			continue
		}

		func() {
			r.completedFadesLock.Lock()
			defer r.completedFadesLock.Unlock()
			r.completedFades[fade.ID] = time.Now()
		}()

		trackfile, ok := r.contacts.getByID(fade.ID)
		if !ok {
			log.Debug().Uint64("id", fade.ID).Msg("faded trackfile not found on scope")
			continue
		}
		log.Info().
			Uint64("id", fade.ID).
			Str("callsign", trackfile.Contact.Name).
			Str("aircraft", trackfile.Contact.ACMIName).
			Stringer("coalition", trackfile.Contact.Coalition).
			Msg("removing faded trackfile")

		isGrouped := false
		for _, grp := range groups {
			if isTrackfileInGroup(trackfile, &grp) {
				isGrouped = true
				break
			}
		}
		if !isGrouped {
			grp := r.findGroupForAircraft(trackfile)
			if grp != nil {
				groups = append(groups, *grp)
			}
		}
	}
	return groups
}

// handleFaded collects faded contacts into groups, removes the contacts from the database, and calls the fadedCallback.
func (r *Radar) handleFaded(fades []sim.Faded) {
	groups := r.collectFadedGroups(fades)

	for _, fade := range fades {
		r.contacts.delete(fade.ID)
	}

	r.callbackLock.RLock()
	defer r.callbackLock.RUnlock()
	if r.fadedCallback == nil {
		return
	}
	for _, grp := range groups {
		r.fadedCallback(grp.point(), &grp, grp.contacts[0].Contact.Coalition)
	}
}

func (r *Radar) areFadesPending() bool {
	r.pendingFadesLock.RLock()
	defer r.pendingFadesLock.RUnlock()
	return len(r.pendingFades) > 0
}

// WaitUntilFadesResolve blocks until all fade events have been processed, or the context is cancelled.
func (r *Radar) WaitUntilFadesResolve(ctx context.Context) {
	if !r.areFadesPending() {
		return
	}
	logger := log.Sample(zerolog.Rarely)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !r.areFadesPending() {
				return
			}
			logger.Debug().Msg("waiting for pending fades to resolve")
		}
	}
}
