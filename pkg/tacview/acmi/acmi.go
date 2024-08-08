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
	// Start should be called before Stream() to initialize the ACMI stream.
	Start(context.Context) error
}

type streamer struct {
	acmi           *bufio.Reader
	referencePoint orb.Point
	referenceTime  time.Time
	cursorTime     time.Time
	objects        map[int]*types.Object
	objectsLock    sync.RWMutex
	fades          chan *types.Object
	bullseyesIdx   sync.Map
	updateInterval time.Duration
	inMultiline    bool
	catchUpCounter int
}

func New(acmi *bufio.Reader, updateInterval time.Duration) ACMI {
	return &streamer{
		acmi:           acmi,
		referencePoint: orb.Point{0, 0},
		referenceTime:  time.Now(),
		cursorTime:     time.Now(),
		objects:        make(map[int]*types.Object),
		fades:          make(chan *types.Object),
		updateInterval: updateInterval,
		objectsLock:    sync.RWMutex{},
		bullseyesIdx:   sync.Map{},
	}
}

func (s *streamer) Start(ctx context.Context) error {
	log.Info().Msg("starting ACMI protocol handler")
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			line, err := s.acmi.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					log.Debug().Int("count", s.catchUpCounter).Msg("caught up to ACMI stream")
					s.catchUpCounter++
					time.Sleep(1 * time.Second)
					if s.catchUpCounter > 120 {
						log.Warn().Int("count", s.catchUpCounter).Msg("!!! SUSPECTED END OF STREAM - POSSIBLE SERVER RESTART !!!")
					}
				} else {
					return fmt.Errorf("error reading ACMI stream: %w", err)
				}
			} else {
				err = s.handleLine(line)
				if err != nil {
					log.Error().Err(err).Str("line", line).Msg("error handling ACMI line")
				}
			}
		}
	}
}

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
		timeframe, err := types.ParseTimeFrame(line)

		if err != nil {
			log.Error().Err(err).Msg("error parsing time frame")
			return fmt.Errorf("error parsing time frame: %w", err)
		}

		s.cursorTime = s.referenceTime.Add(timeframe.Offset)
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
			logger.Trace().Time("referenceTime", s.referenceTime).Msg("reference time updated")
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

	logger = logger.With().Int("id", update.ID).Logger()

	s.objectsLock.Lock()
	defer s.objectsLock.Unlock()
	if update.IsRemoval {
		object, ok := s.objects[update.ID]
		if ok {
			s.fades <- object
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

func (s *streamer) Stream(ctx context.Context, updates chan<- sim.Updated, fades chan<- sim.Faded) {
	ticker := time.NewTicker(s.updateInterval)
	defer ticker.Stop()
	s.processUpdates(updates)
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping ACMI stream due to context cancellation")
			return
		case object := <-s.fades:
			fades <- sim.Faded{
				Timestamp: time.Now(),
				UnitID:    uint32(object.ID),
			}
		case <-ticker.C:
			s.processUpdates(updates)
		}
	}
}

func (s *streamer) processUpdates(updates chan<- sim.Updated) {
	s.objectsLock.Lock()
	defer s.objectsLock.Unlock()
	for _, object := range s.objects {
		logger := log.With().Int("id", object.ID).Logger()
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

func (s *streamer) updateAircraft(updates chan<- sim.Updated, object *types.Object) error {
	logger := log.With().Int("id", object.ID).Logger()

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

func (s *streamer) updateBullseye(object *types.Object) {
	logger := log.With().Int("id", object.ID).Logger()
	prop, ok := object.GetProperty(properties.Coalition)
	if !ok {
		logger.Warn().Msg("bullseye has no coalition")
		return
	}
	coalition := properties.PropertyToCoalition(prop)
	s.bullseyesIdx.Store(coalition, object.ID)
}

func (s *streamer) Bullseye(coalition coalitions.Coalition) (p orb.Point) {
	val, ok := s.bullseyesIdx.Load(coalition)
	if !ok {
		return
	}
	objectID := val.(int)
	s.objectsLock.RLock()
	defer s.objectsLock.RUnlock()
	object, ok := s.objects[objectID]
	if !ok {
		return
	}
	coordinates, err := object.GetCoordinates(s.referencePoint)
	if err != nil {
		return
	}
	return coordinates.Location

}

func (s *streamer) Time() time.Time {
	return s.cursorTime
}

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
			UnitID:    uint32(object.ID),
			Name:      callsign,
			Coalition: coalition,
			ACMIName:  name,
		},
		Frame: frame,
	}, nil
}
