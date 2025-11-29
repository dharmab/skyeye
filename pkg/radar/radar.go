// Package radar implements mid-level logic for Ground-Controlled Interception (GCI).
package radar

import (
	"context"
	"sync"
	"time"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/encyclopedia"
	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/rs/zerolog/log"
)

// Radar consumes updates from the simulation, keeps track of each aircraft as a trackfile, and provides functions to collect the aircraft into groups.
type Radar struct {
	// starts receives an event whenever a mission (re)starts.
	starts <-chan sim.Started
	// updates receives frames updating the position of aircraft.
	updates <-chan sim.Updated
	// fades receives events when aircraft are marked as removed.
	fades <-chan sim.Faded
	// missionTime should be continually updated to the current mission time.
	missionTime time.Time
	// missionTimeLock protects missionTime.
	missionTimeLock sync.RWMutex
	// coalition is the player coalition.
	coalition coalitions.Coalition
	// bullsyses maps coalitions to their respective bullseye points.
	bullseyes sync.Map
	// contacts contains trackfiles for each aircraft.
	contacts *contactDatabase
	// startedCallback is called when a start event is received.
	startedCallback StartedCallback
	// fadedCallback is called when a fade event is received.
	fadedCallback FadedCallback
	// removalCallback is called when a trackfile is removed for a reason other than a fade event.
	removalCallback RemovedCallback
	// callbackLock protects startedCallback, fadedCallback, and removalCallback.
	callbackLock sync.RWMutex
	// center is a point used to center PICTURE calls.
	center orb.Point
	// centerLock protects center.
	centerLock sync.RWMutex
	// mandatoryThreatRadius is the radius within which a hostile aircraft is always considered a threat.
	mandatoryThreatRadius unit.Length
	// completedFades records the IDs of contacts that have been faded.
	completedFades map[uint64]time.Time
	// completedFadesLock protects completedFades.
	completedFadesLock sync.RWMutex
	// pendingFades collects faded contacts for grouping.
	pendingFades []sim.Faded
	// pendingFadesLock protects pendingFades.
	pendingFadesLock sync.RWMutex
}

// New creates a radar scope that consumes updates from the provided channels.
func New(coalition coalitions.Coalition, starts <-chan sim.Started, updates <-chan sim.Updated, fades <-chan sim.Faded, mandatoryThreatRadius unit.Length) *Radar {
	return &Radar{
		coalition:             coalition,
		starts:                starts,
		updates:               updates,
		fades:                 fades,
		contacts:              newContactDatabase(),
		mandatoryThreatRadius: mandatoryThreatRadius,
		completedFades:        map[uint64]time.Time{},
		pendingFades:          []sim.Faded{},
	}
}

// SetMissionTime updates the mission time. The mission time is used for computing magnetic declination.
func (r *Radar) SetMissionTime(t time.Time) {
	r.missionTimeLock.Lock()
	defer r.missionTimeLock.Unlock()
	r.missionTime = t
}

// SetBullseye updates the bullseye point for the given coalition.
// The bullseye point is the reference point for polar coordinates provided in [Group.Bullseye].
func (r *Radar) SetBullseye(bullseye orb.Point, coalition coalitions.Coalition) {
	current := r.Bullseye(coalition)
	if current.Lon() != bullseye.Lon() || current.Lat() != bullseye.Lat() {
		log.Info().
			Int("coalitionID", int(coalition)).
			Float64("lon", bullseye.Lon()).
			Float64("lat", bullseye.Lat()).
			Msg("updating bullseye")
	}
	r.bullseyes.Store(coalition, bullseye)
}

// Bullseye returns the bullseye point for the given coalition.
func (r *Radar) Bullseye(coalition coalitions.Coalition) orb.Point {
	p, ok := r.bullseyes.Load(coalition)
	if !ok {
		return orb.Point{}
	}
	return p.(orb.Point)
}

// Run consumes updates from the simulation channels until the context is cancelled.
func (r *Radar) Run(ctx context.Context, wg *sync.WaitGroup) {
	wg.Go(func() {
		for {
			ticker := time.NewTicker(5 * time.Second)
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.updateCenterPoint()
			}
		}
	})

	wg.Go(func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-r.starts:
				r.handleStarted()
			}
		}
	})

	wg.Go(func() {
		for {
			select {
			case <-ctx.Done():
				return
			case update := <-r.updates:
				r.handleUpdate(update)
			}
		}
	})
	wg.Go(func() {
		r.collectFadedTrackfiles(ctx)
	})

	wg.Go(func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.handleGarbageCollection()
			}
		}
	})

	<-ctx.Done()
}

// handleUpdate updates the database using the provided update.
func (r *Radar) handleUpdate(update sim.Updated) {
	logger := log.With().
		Str("callsign", update.Labels.Name).
		Str("aircraft", update.Labels.ACMIName).
		Uint64("id", update.Labels.ID).
		Stringer("coalition", update.Labels.Coalition).
		Logger()

	trackfile, ok := r.contacts.getByID(update.Labels.ID)
	if ok {
		trackfile.Update(update.Frame)
	} else {
		trackfile = trackfiles.New(update.Labels)
		r.contacts.set(trackfile)
		logger.Info().Msg("created new trackfile")
	}
}

// handleGarbageCollection removes trackfiles that have not been updated in a long time.
func (r *Radar) handleGarbageCollection() {
	r.pendingFadesLock.RLock()
	defer r.pendingFadesLock.RUnlock()
	r.missionTimeLock.RLock()
	defer r.missionTimeLock.RUnlock()
	if len(r.pendingFades) > 0 {
		return
	}

	for trackfile := range r.contacts.values() {
		logger := log.With().
			Uint64("id", trackfile.Contact.ID).
			Str("callsign", trackfile.Contact.Name).
			Str("aircraft", trackfile.Contact.ACMIName).
			Stringer("coalition", trackfile.Contact.Coalition).
			Logger()

		lastSeen := trackfile.LastKnown().Time
		isOld := lastSeen.Before(r.missionTime.Add(-1 * time.Minute))
		if !lastSeen.IsZero() && isOld {
			ok := r.contacts.delete(trackfile.Contact.ID)
			if ok {
				logger.Info().
					Stringer("age", r.missionTime.Sub(lastSeen)).
					Msg("expired trackfile")
				go func() {
					r.callbackLock.RLock()
					defer r.callbackLock.RUnlock()
					if r.removalCallback != nil {
						r.removalCallback(trackfile)
					}
				}()
			}
		}
	}
}

// isValidTrack checks if the trackfile is valid. This means the following conditions are met:
//   - Last known position is not (0, 0)
//   - If AGL is known, AGL is above 10 meters
//   - If AGL is unknown, speed is above 50 knots
func isValidTrack(trackfile *trackfiles.Trackfile) bool {
	if trackfile.IsLastKnownPointZero() {
		return false
	}

	agl := trackfile.LastKnown().AGL
	if agl != nil {
		return *agl > 10*unit.Meter
	}

	return trackfile.Speed() > 50*unit.Knot
}

// isMatch checks:
//   - if the trackfile is of the given coalition
//   - if the trackfile is of the given contact category (or if the aircraft is not in the encyclopedia)
//   - if the trackfile is valid
func isMatch(trackfile *trackfiles.Trackfile, coalition coalitions.Coalition, filter brevity.ContactCategory) bool {
	if trackfile.Contact.Coalition != coalition {
		return false
	}
	if !isValidTrack(trackfile) {
		return false
	}
	data, ok := encyclopedia.GetAircraftData(trackfile.Contact.ACMIName)
	// If the aircraft is not in the encyclopedia, assume it matches
	matchesFilter := !ok || data.Category() == filter || filter == brevity.Aircraft
	return matchesFilter
}

// Declination returns the magnetic declination at the given point, at the time provided in SetMissionTime.
func (r *Radar) Declination(p orb.Point) unit.Angle {
	r.missionTimeLock.RLock()
	defer r.missionTimeLock.RUnlock()
	declination, err := bearings.Declination(p, r.missionTime)
	log.Debug().Any("declination", declination).Msgf("computed magnetic radar declination at point %v", p)

	if err != nil {
		log.Error().Err(err).Msg("failed to get declination")
	}
	return declination
}
