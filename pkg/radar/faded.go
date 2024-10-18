package radar

import (
	"context"
	"time"

	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/dharmab/skyeye/pkg/trackfiles"
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
func (s *scope) collectFaded(ctx context.Context) {
	// Whenenver we pass the deadline, we collect the faded contacts into groups and call the fadedCallback.
	var deadline time.Time

	// We check the deadline at intervals.
	ticker := time.NewTicker(10 * time.Second)

	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case fade := <-s.fades:
			// When we receive a faded contact, we wait a little in case it's wingman is also fading.
			// This is common if the flight lands or is being engaged by a coordinated flight.
			deadline = time.Now().Add(15 * time.Second)
			go func() {
				s.pendingFadesLock.Lock()
				defer s.pendingFadesLock.Unlock()
				s.pendingFades = append(s.pendingFades, fade)
			}()
		case <-ticker.C:
			go func() {
				s.pendingFadesLock.Lock()
				defer s.pendingFadesLock.Unlock()
				if len(s.pendingFades) > 0 && time.Now().After(deadline) {
					// The fade events have settled down now
					s.handleFaded(s.pendingFades)
					s.pendingFades = []sim.Faded{}
				}
			}()
		}
	}
}

// handleFaded collects faded contacts into groups, removes the contacts from the database, and calls the fadedCallback.
func (s *scope) handleFaded(fades []sim.Faded) {
	var groups []group
	for _, fade := range fades {
		// Find the trackfile for the faded contact
		trackfile, ok := s.contacts.getByID(fade.ID)
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
			grp := s.findGroupForAircraft(trackfile)
			if grp != nil {
				groups = append(groups, *grp)
			}
		}
	}

	// remove the faded contacts from the database
	for _, fade := range fades {
		s.contacts.delete(fade.ID)
	}

	// call the faded callback for each group
	s.callbackLock.RLock()
	defer s.callbackLock.RUnlock()
	for _, grp := range groups {
		if s.fadedCallback != nil {
			s.fadedCallback(grp.point(), &grp, grp.contacts[0].Contact.Coalition)
		}
	}
}
