package radar

import (
	"context"
	"time"

	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/dharmab/skyeye/pkg/trackfiles"
)

func isTrackfileInGroup(trackfile *trackfiles.Trackfile, grp *group) bool {
	for _, contact := range grp.contacts {
		if contact.Contact.UnitID == trackfile.Contact.UnitID {
			return true
		}
	}
	return false
}

// collectFaded continuously collects faded contacts. When there is no new faded contact for 10 seconds,
// it collects all faded contacts into groups, removes the contacts from the database, and calls the fadedCallback.
func (s *scope) collectFaded(ctx context.Context) {
	collectedFades := []sim.Faded{}

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
			deadline = time.Now().Add(10 * time.Second)
			collectedFades = append(collectedFades, fade)
		case <-ticker.C:
			if len(collectedFades) > 0 && time.Now().After(deadline) {
				// The fade events have settled down now
				s.handleFaded(collectedFades)
				collectedFades = []sim.Faded{}
			}
		}
	}
}

// handleFaded collects faded contacts into groups, removes the contacts from the database, and calls the fadedCallback.
func (s *scope) handleFaded(fades []sim.Faded) {
	var groups []group
	for _, fade := range fades {
		// Find the trackfile for the faded contact
		trackfile, ok := s.contacts.getByUnitID(fade.UnitID)
		if !ok {
			continue
		}

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
		s.contacts.delete(fade.UnitID)
	}

	// call the faded callback for each group
	for _, grp := range groups {
		if s.fadedCallback != nil {
			s.fadedCallback(&grp, grp.contacts[0].Contact.Coalition)
		}
	}
}
