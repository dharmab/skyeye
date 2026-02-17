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
	"github.com/dharmab/skyeye/pkg/encyclopedia/terrains"
	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/dharmab/skyeye/pkg/spatial/projections"
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
	// enableTerrainDetection controls whether terrain detection and Transverse Mercator projection are used.
	// When false, spatial functions use spherical Earth calculations.
	enableTerrainDetection bool
	// projection is a best guess at the current map projection based on the bullseye.
	// This improves accuracy for distance/bearing calculations at extreme latitudes.
	// If it is nil, either because terrain detection is disabled or because no bullseye has been set,
	// spatial functions fall back to spherical Earth calculations.
	projection projections.Projection
	// projectionLock protects projection.
	projectionLock sync.RWMutex
}

// New creates a radar scope that consumes updates from the provided channels.
// When enableTerrainDetection is true, SetBullseye will detect the closest DCS terrain and use its
// Transverse Mercator projection for spatial calculations. When false, spherical Earth calculations are used.
func New(coalition coalitions.Coalition, starts <-chan sim.Started, updates <-chan sim.Updated, fades <-chan sim.Faded, mandatoryThreatRadius unit.Length, enableTerrainDetection bool) *Radar {
	return &Radar{
		coalition:              coalition,
		starts:                 starts,
		updates:                updates,
		fades:                  fades,
		contacts:               newContactDatabase(),
		mandatoryThreatRadius:  mandatoryThreatRadius,
		enableTerrainDetection: enableTerrainDetection,
		completedFades:         map[uint64]time.Time{},
		pendingFades:           []sim.Faded{},
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
// When terrain detection is enabled, this also detects the closest DCS terrain and updates
// the Transverse Mercator projection used for spatial calculations.
func (r *Radar) SetBullseye(bullseye orb.Point, coalition coalitions.Coalition) {
	current := r.Bullseye(coalition)
	if current.Lon() != bullseye.Lon() || current.Lat() != bullseye.Lat() {
		if r.enableTerrainDetection {
			terrain := terrains.Closest(bullseye)
			log.Info().
				Int("coalitionID", int(coalition)).
				Float64("lon", bullseye.Lon()).
				Float64("lat", bullseye.Lat()).
				Str("terrain", terrain.Name).
				Msg("updating bullseye")

			r.projectionLock.Lock()
			r.projection = terrain.Projection()
			r.projectionLock.Unlock()
		} else {
			log.Info().
				Int("coalitionID", int(coalition)).
				Float64("lon", bullseye.Lon()).
				Float64("lat", bullseye.Lat()).
				Msg("updating bullseye")
		}
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

// Projection returns the current Transverse Mercator projection.
// This projection should be used for distance and bearing calculations to improve
// accuracy at extreme latitudes. Returns nil if terrain detection is disabled or
// no projection has been set. Callers should handle a nil return by falling back
// to spherical Earth calculations.
func (r *Radar) Projection() projections.Projection {
	r.projectionLock.RLock()
	defer r.projectionLock.RUnlock()
	return r.projection
}

// withProjection returns a spatial.Option that uses the current projection.
// This is a convenience helper for passing to spatial functions.
// When the projection is nil (terrain detection disabled or no bullseye set),
// spatial functions fall back to spherical Earth calculations.
func (r *Radar) withProjection() spatial.Option {
	return spatial.WithProjection(r.Projection())
}

// setBullseyeForGroup computes and sets the bullseye for a group using the current projection.
func (r *Radar) setBullseyeForGroup(grp *group) {
	bullseyePoint := r.Bullseye(r.coalition)
	if spatial.IsZero(bullseyePoint) {
		return
	}
	groupPoint := grp.point()
	declination := r.Declination(groupPoint)
	bearing := spatial.TrueBearing(bullseyePoint, groupPoint, r.withProjection()).Magnetic(declination)
	distance := spatial.Distance(bullseyePoint, groupPoint, r.withProjection())
	grp.bullseye = brevity.NewBullseye(bearing, distance)
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
	if err != nil {
		log.Error().Err(err).Msg("failed to get declination")
	}
	return declination
}
