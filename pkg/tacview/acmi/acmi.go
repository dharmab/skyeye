// package acmi streams simulation data from a TacView Air Combat Maneuvering
// Instrumentation (ACMI) data source.
package acmi

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/dharmab/skyeye/pkg/tacview/properties"
	"github.com/dharmab/skyeye/pkg/tacview/tags"
	"github.com/dharmab/skyeye/pkg/tacview/types"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/paulmach/orb"
	"github.com/rs/zerolog/log"
)

// https://www.tacview.net/documentation/acmi/en/

// ACMI is an interface for streaming simulation data from a Tacview ACMI data source.
type ACMI interface {
	sim.Sim
	// Run should be called to stream ACMI data. It may be called multiple times,
	// but should not be called concurrently. It may return an error that wraps io.EOF
	// to indicate the end of the ACMI data stream, which may occur when the sim has restarted.
	// If that occurs, recovery is usually possible by restarting the stream.
	Run(context.Context) error
}

type streamer struct {
	// acmi reads ACMI data lines.
	acmi *bufio.Reader

	// referencePoint is a center point that must be added to the coordinates of all objects to get the true coordinates.
	referencePoint orb.Point
	// referenceTime is the base time for the current mission. It must be added to the frame offset time of each event to get the mission time.
	referenceTime time.Time
	// cursorTime is the mission time of the frame currently being processed.
	cursorTime time.Time
	// objects maps object IDs to data.
	objects map[uint64]*types.Object
	// objectsLock protects the objects map.
	objectsLock sync.RWMutex
	// starts is an internal channel for passing the real-time observed time when the mission starts (i.e. the reference time is first set)
	starts chan time.Time
	// started tracks if a start event has been published
	started bool
	// removals is an internal channel for passing messages when objects are removed.
	removals chan *types.Object
	// bullseyesIdx indexes bullseye object IDs by coalition.
	bullseyesIdx sync.Map
	// updateInterval is the interval at which the streamer will publish object updates.s
	updateInterval time.Duration
	// inMultiline is true when the streamer is currently processing a line that contains newline characters.
	inMultiline bool
	// eofCounter is incremented each time an EOF is received.
	eofCounter int
}

// New creates a new ACMI streamer. The ACMI data is read from the provided reader. The updateInterval
// is the interval at which the streamer will publish to the updates channel. The endDelay is the
// duration to wait after the last ACMI data is read before considering the stream to have ended.
func New(acmi *bufio.Reader, updateInterval time.Duration) ACMI {
	return &streamer{
		acmi:           acmi,
		objects:        make(map[uint64]*types.Object),
		starts:         make(chan time.Time),
		removals:       make(chan *types.Object),
		updateInterval: updateInterval,
	}
}

// Run implements [ACMI.Run]. It reads lines from the ACMI data source and handles them one at a time.
// If no lines are read for a grace period, the stream is considered to have ended and Run returns.
func (s *streamer) Run(ctx context.Context) error {
	log.Info().Msg("starting ACMI protocol handler")
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if s.eofCounter > 30 {
				log.Warn().Msg("Received EOF 30 times, suspected server restart, stopping ACMI stream")
				return io.EOF
			}

			line, err := s.acmi.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					s.eofCounter++
				} else {
					return fmt.Errorf("error reading ACMI stream: %w", err)
				}
			}

			err = s.handleLine(line)
			if err != nil {
				log.Error().Err(err).Str("line", line).Msg("error handling ACMI line")
			}
		}
	}
}

// handleLine parses a line of ACMI data.
//   - headers and comments are ignored.
//   - global object updates update the reference time and point.
//   - object updates are stored in the object map.
//   - object removals remove the object from the internal object map and publish to the removals channel.
func (s *streamer) handleLine(line string) error {
	if strings.HasSuffix(line, "\\\n") {
		s.inMultiline = true
	}
	if s.inMultiline {
		if !strings.HasSuffix(line, "\\\n") {
			s.inMultiline = false
		}
		log.Debug().Str("line", line).Msg("skipping multiline line")
		return nil
	}

	line = strings.TrimSpace(line)
	if line == "" {
		log.Trace().Msg("line is empty")
		return nil
	}
	logger := log.With().Str("line", line).Logger()

	// Comments
	if strings.HasPrefix(line, "//") {
		logger.Debug().Msg("line is a comment")
		return nil
	}

	// Headers
	if line == fmt.Sprintf("%s=%s", properties.FileType, properties.FileTypeTacView) {
		log.Debug().Msg("ACMI file type detected")
		return nil
	}
	if line == fmt.Sprintf("%s=%s", properties.FileVersion, properties.FileVersion2_2) {
		log.Debug().Msg("ACMI flight recording version 2.2 detected")
		return nil
	}

	// A line beginning with a # is a new relative time frame, relative to the global object's reference time.
	if line[0] == '#' {
		offset, err := types.ParseTimeFrame(line)
		if err != nil {
			log.Error().Err(err).Msg("error parsing time frame")
			return fmt.Errorf("error parsing time frame: %w", err)
		}

		s.cursorTime = s.referenceTime.Add(offset)
		return nil
	}

	// Object updates
	update, err := types.ParseObjectUpdate(line)
	if err != nil {
		logger.Error().Err(err).Msg("error parsing object update")
		return fmt.Errorf("error parsing object update: %w", err)
	}

	if update.IsGlobal {
		var updateErr error
		if _, ok := update.Properties[properties.ReferenceTime]; ok {
			referenceTime, err := time.Parse(time.RFC3339, update.Properties[properties.ReferenceTime])
			if err != nil {
				logger.Error().Err(err).Msg("error parsing reference time")
				updateErr = errors.Join(updateErr, fmt.Errorf("error parsing reference time: %w", err))
			}
			s.referenceTime = referenceTime
			logger.Debug().Time("referenceTime", s.referenceTime).Msg("reference time updated")
			if !s.started {
				s.starts <- time.Now()
				s.started = true
			}
		}
		if _, ok := update.Properties[properties.ReferenceLongitude]; ok {
			longitude, err := strconv.ParseFloat(update.Properties[properties.ReferenceLongitude], 64)
			if err != nil {
				logger.Error().Err(err).Msg("error parsing reference longitude")
				updateErr = errors.Join(updateErr, fmt.Errorf("error parsing reference longitude: %w", err))
			}
			s.referencePoint = orb.Point{longitude, s.referencePoint.Lat()}
			logger.Trace().Float64("longitude", longitude).Msg("reference point updated")
		}
		if _, ok := update.Properties[properties.ReferenceLatitude]; ok {
			latitude, err := strconv.ParseFloat(update.Properties[properties.ReferenceLatitude], 64)
			if err != nil {
				logger.Error().Err(err).Msg("error parsing reference latitude")
				updateErr = errors.Join(updateErr, fmt.Errorf("error parsing reference latitude: %w", err))
			}
			s.referencePoint = orb.Point{s.referencePoint.Lon(), latitude}
			logger.Trace().Float64("latitude", latitude).Msg("reference point updated")
		}
		if updateErr != nil {
			return fmt.Errorf("error updating global object: %w", updateErr)
		}
		return nil
	}

	logger = logger.With().Uint64("id", update.ID).Logger()

	s.objectsLock.Lock()
	defer s.objectsLock.Unlock()
	if update.IsRemoval {
		object, ok := s.objects[update.ID]
		if ok {
			s.removals <- object
			delete(s.objects, update.ID)
		}
		return nil
	}

	if _, ok := s.objects[update.ID]; !ok {
		s.objects[update.ID] = types.NewObject(update.ID)
	}
	for k, v := range update.Properties {
		s.objects[update.ID].SetProperty(k, v)
	}
	return nil
}

// Stream implements [ACMI.Stream].
func (s *streamer) Stream(ctx context.Context, starts chan<- sim.Started, updates chan<- sim.Updated, fades chan<- sim.Faded) {
	ticker := time.NewTicker(s.updateInterval)
	defer ticker.Stop()
	s.processUpdates(updates)
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping ACMI stream due to context cancellation")
			return
		case object := <-s.removals:
			fades <- sim.Faded{
				Timestamp: time.Now(),
				ID:        object.ID,
			}
		case observedAt := <-s.starts:
			starts <- sim.Started{
				Timestamp:        observedAt,
				MissionTimestamp: s.referenceTime,
			}
		case <-ticker.C:
			s.processUpdates(updates)
		}
	}
}

// processUpdates publishes updates for all objects.
func (s *streamer) processUpdates(updates chan<- sim.Updated) {
	s.objectsLock.Lock()
	defer s.objectsLock.Unlock()
	for _, object := range s.objects {
		logger := log.With().Uint64("id", object.ID).Logger()
		types, err := object.GetTypes()
		if err != nil {
			logger.Error().Err(err).Msg("error getting object types")
			continue
		}

		if slices.Contains(types, tags.Bullseye) {
			s.updateBullseye(object)
		}
		if slices.Contains(types, tags.FixedWing) || slices.Contains(types, tags.Rotorcraft) {
			if err := s.updateAircraft(updates, object); err != nil {
				logger.Error().Err(err).Msg("error updating aircraft")
				continue
			}
		}
	}
}

// updateAircraft publishes an update for an aircraft object. It wraps buildUpdate with logging and error handling.
func (s *streamer) updateAircraft(updates chan<- sim.Updated, object *types.Object) error {
	logger := log.With().Uint64("id", object.ID).Logger()

	update, err := s.buildUpdate(object)
	if err != nil {
		logger.Error().Err(err).Msg("error building object update")
		return err
	}
	if update != nil {
		updates <- *update
	}
	return nil
}

// updateBullseye indexes the given bullseye object in the bullseyes index.
func (s *streamer) updateBullseye(object *types.Object) {
	logger := log.With().Uint64("id", object.ID).Logger()
	prop, ok := object.GetProperty(properties.Coalition)
	if !ok {
		logger.Warn().Msg("bullseye has no coalition")
		return
	}
	coalition := properties.PropertyToCoalition(prop)
	s.bullseyesIdx.Store(coalition, object.ID)
}

// Bullseye implements [ACMI.Bullseye].
func (s *streamer) Bullseye(coalition coalitions.Coalition) (orb.Point, error) {
	val, ok := s.bullseyesIdx.Load(coalition)
	if !ok {
		return orb.Point{}, errors.New("bullseye for coalition not found")
	}
	objectID := val.(uint64)
	s.objectsLock.RLock()
	defer s.objectsLock.RUnlock()
	object, ok := s.objects[objectID]
	if !ok {
		return orb.Point{}, errors.New("bullseye object for coalition not found")
	}
	coordinates, err := object.GetCoordinates(s.referencePoint)
	if err != nil {
		return orb.Point{}, fmt.Errorf("error getting bullseye coordinates: %w", err)
	}
	return coordinates.Location, nil
}

// Time implements [ACMI.Time].
func (s *streamer) Time() time.Time {
	return s.cursorTime
}

// buildUpdate creates an aircraft update from an object.
func (s *streamer) buildUpdate(object *types.Object) (*sim.Updated, error) {
	types, err := object.GetTypes()
	if err != nil {
		return nil, fmt.Errorf("error getting object types: %w", err)
	}

	if !slices.Contains(types, tags.FixedWing) && !slices.Contains(types, tags.Rotorcraft) {
		return nil, errors.New("object is not an aircraft")
	}
	name, ok := object.GetProperty(properties.Name)
	if !ok {
		return nil, errors.New("object has no name")
	}
	coordinates, err := object.GetCoordinates(s.referencePoint)
	if err != nil {
		return nil, err
	}
	if coordinates == nil {
		return nil, nil
	}

	acmiCoalition, ok := object.GetProperty(properties.Coalition)
	if !ok {
		return nil, errors.New("object has no coalition")
	}

	coalition := properties.PropertyToCoalition(acmiCoalition)

	callsign, ok := object.GetProperty(properties.Pilot)
	if !ok {
		// If the object has no pilot, use the object ID as the callsign.
		callsign = fmt.Sprintf("Unit %d", object.ID)
	}

	frame := trackfiles.Frame{
		Time:     s.cursorTime,
		Point:    coordinates.Location,
		Altitude: coordinates.Altitude,
		Heading:  coordinates.Heading,
	}

	return &sim.Updated{
		Labels: trackfiles.Labels{
			ID:        object.ID,
			Name:      callsign,
			Coalition: coalition,
			ACMIName:  name,
		},
		Frame: frame,
	}, nil
}
