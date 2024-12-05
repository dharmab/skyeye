package radar

import (
	"context"
	"time"

	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func isTrackfileInGroup(trackfile *trackfiles.Trackfile, grp *group) bool {
	for _, contact := range grp.contacts {
		if contact.Contact.ID == trackfile.Contact.ID {
			return true
		}
	}
	return false
}

// collectFaded continuously collects faded contacts. When there is no new faded contact for 10 seconds,
// it collects all faded contacts into groups, removes the contacts from the database, and calls the fadedCallback.
func (r *Radar) collectFaded(ctx context.Context) {
	// Whenenver we pass the deadline, we collect the faded contacts into groups and call the fadedCallback.
	var deadline time.Time

	// We check the deadline at intervals.
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
			go func() {
				r.pendingFadesLock.Lock()
				defer r.pendingFadesLock.Unlock()
				if len(r.pendingFades) > 0 && time.Now().After(deadline) {
					// The fade events have settled down now
					r.handleFaded(r.pendingFades)
					r.pendingFades = []sim.Faded{}
				}
			}()
		}
	}
}

// handleFaded collects faded contacts into groups, removes the contacts from the database, and calls the fadedCallback.
func (r *Radar) handleFaded(fades []sim.Faded) {
	var groups []group
	for _, fade := range fades {
		// Find the trackfile for the faded contact
		trackfile, ok := r.contacts.getByID(fade.ID)
		if !ok {
			continue
		}

		log.Info().
			Uint64("id", fade.ID).
			Str("callsign", trackfile.Contact.Name).
			Str("aircraft", trackfile.Contact.ACMIName).
			Stringer("coalition", trackfile.Contact.Coalition).
			Msg("removing trackfile")

		// Check if the trackfile is already collected into a group
		isGrouped := false
		for _, grp := range groups {
			if isTrackfileInGroup(trackfile, &grp) {
				isGrouped = true
				break
			}
		}
		// If the trackfile is not already collected into a group, create a new group
		if !isGrouped {
			grp := r.findGroupForAircraft(trackfile)
			if grp != nil {
				groups = append(groups, *grp)
			}
		}
	}

	// remove the faded contacts from the database
	for _, fade := range fades {
		r.contacts.delete(fade.ID)
	}

	// call the faded callback for each group
	r.callbackLock.RLock()
	defer r.callbackLock.RUnlock()
	for _, grp := range groups {
		if r.fadedCallback != nil {
			r.fadedCallback(grp.point(), &grp, grp.contacts[0].Contact.Coalition)
		}
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
