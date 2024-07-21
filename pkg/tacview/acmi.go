package tacview

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/dharmab/skyeye/pkg/tacview/properties"
	"github.com/dharmab/skyeye/pkg/tacview/tags"
	"github.com/dharmab/skyeye/pkg/tacview/types"
	"github.com/dharmab/skyeye/pkg/trackfile"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/rs/zerolog/log"
)

// https://www.tacview.net/documentation/acmi/en/

type ACMI interface {
	sim.Sim
	Start(context.Context) error
}

type streamer struct {
	acmi           *bufio.Reader
	referencePoint orb.Point
	referenceTime  time.Time
	cursorTime     time.Time
	objects        map[int]types.Object
	fades          chan types.Object
	bullseye       orb.Point
}

func NewACMI(coalition coalitions.Coalition, acmi *bufio.Reader) ACMI {
	return &streamer{
		acmi:           acmi,
		referencePoint: orb.Point{0, 0},
		referenceTime:  time.Now(),
		cursorTime:     time.Now(),
		objects:        make(map[int]types.Object),
		fades:          make(chan types.Object),
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
					log.Debug().Msg("caught up to ACMI stream")
					time.Sleep(1 * time.Second)
				} else {
					return fmt.Errorf("error reading ACMI stream: %w", err)
				}
			} else {
				log.Trace().Str("line", line).Msg("handling ACMI line")
				err = s.handleLine(line)
				if err != nil {
					log.Error().Err(err).Str("line", line).Msg("error handling ACMI line")
				}
			}
		}
	}
}

func (s *streamer) handleLine(line string) error {
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

		logger.Trace().Msg("line is a new relative time frame")
		s.cursorTime = s.referenceTime.Add(timeframe.Offset)
		log.Trace().Dur("offset", timeframe.Offset).Time("cursor", s.cursorTime).Msg("relative time updated")
		return nil
	}

	// Object updates
	update, err := types.ParseObjectUpdate(line)
	if err != nil {
		logger.Error().Err(err).Msg("error parsing object update")
		return fmt.Errorf("error parsing object update: %w", err)
	}

	if update.IsGlobal {
		logger.Trace().Msg("line is a global object update")
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

	if update.IsRemoval {
		logger.Trace().Msg("line is an object removal")
		object, ok := s.objects[update.ID]
		if ok {
			logger.Trace().Msg("publishing object to fade channel")
			s.fades <- object
			logger.Trace().Msg("removing object from map")
			delete(s.objects, update.ID)
			logger.Trace().Msg("object removed")
		} else {
			logger.Trace().Msg("object not found in map")
		}
		return nil
	}

	if _, ok := s.objects[update.ID]; !ok {
		logger.Trace().Msg("watching new object")
		s.objects[update.ID] = types.Object{ID: update.ID, Properties: make(map[string]string)}
	}
	logger.Trace().Msg("updating object properties")
	for k, v := range update.Properties {
		s.objects[update.ID].Properties[k] = v
		logger.Trace().Str("name", k).Str("value", v).Msg("object property updated")
	}
	return nil
}

func (s *streamer) Stream(ctx context.Context, updates chan<- sim.Updated, fades chan<- sim.Faded) {
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping ACMI stream due to context cancellation")
			return
		case object := <-s.fades:
			log.Trace().Int("id", object.ID).Msg("object faded")
			fades <- sim.Faded{
				Timestamp: time.Now(),
				UnitID:    uint32(object.ID),
			}
		case <-ticker.C:
			log.Debug().Msg("iterating over objects for trackfile updates")
			for _, object := range s.objects {
				logger := log.With().Int("id", object.ID).Logger()
				types, err := object.GetTypes()
				if err != nil {
					logger.Error().Err(err).Msg("error getting object types")
					continue
				}
				logger.Trace().Interface("types", types).Msg("checking object types")

				if slices.Contains(types, tags.Bullseye) {
					logger.Trace().Msg("object is bullseye")
					// TODO check coalition
					err := s.updateBullseye(object)
					if err != nil {
						logger.Error().Err(err).Msg("error updating bullseye")
						continue
					}
					logger.Info().Msg("bullseye updated")
				}
				if slices.Contains(types, tags.FixedWing) || slices.Contains(types, tags.Rotorcraft) {
					logger.Trace().Msg("object is an aircraft")
					coordinates, err := object.GetCoordinates(s.referencePoint)
					if err != nil {
						logger.Error().Err(err).Msg("error getting object coordinates")
						continue
					}
					if coordinates.Altitude < unit.Length(10)*unit.Meter {
						logger.Trace().Float64("agl", coordinates.Altitude.Meters()).Msg("object is below altitude threshold")
						continue
					}

					update, err := s.buildUpdate(object)
					if err != nil {
						logger.Error().Err(err).Msg("error building object update")
						continue
					}
					logger.Info().Int("unitID", int(update.Aircraft.UnitID)).Str("name", update.Aircraft.Name).Str("aircraft", update.Aircraft.ACMIName).Msg("aircraft update")
					updates <- *update
				}
			}
		}
	}
}

func (s *streamer) Bullseye() orb.Point {
	return s.bullseye
}

func (s *streamer) updateBullseye(object types.Object) error {
	types, err := object.GetTypes()
	if err != nil {
		return err
	}
	if !slices.Contains(types, tags.Bullseye) {
		return nil
	}
	coordinates, err := object.GetCoordinates(s.referencePoint)
	if err != nil {
		return err
	}
	s.bullseye = coordinates.Location
	return nil
}

func (s *streamer) buildUpdate(object types.Object) (*sim.Updated, error) {
	types, err := object.GetTypes()
	if err != nil {
		return nil, err
	}
	if !slices.Contains(types, tags.FixedWing) && !slices.Contains(types, tags.Rotorcraft) {
		return nil, errors.New("object is not an aircraft")
	}
	name, err := object.GetProperty(properties.Name)
	if err != nil {
		return nil, err
	}
	coordinates, err := object.GetCoordinates(s.referencePoint)
	if err != nil {
		return nil, err
	}
	airspeed, err := object.GetSpeed(properties.TAS)
	if err != nil {
		airspeed = 0
	}
	acmiCoalition, err := object.GetProperty(properties.Coalition)
	if err != nil {
		return nil, err
	}
	var coalition coalitions.Coalition
	// Red = Allies because DCS descends from Flanker
	if acmiCoalition == properties.AlliesCoalition {
		coalition = coalitions.Red
	} else if acmiCoalition == properties.EnemiesCoalition {
		coalition = coalitions.Blue
	} else {
		coalition = coalitions.Neutrals
	}

	callsign, err := object.GetProperty(properties.Pilot)
	if err != nil {
		log.Warn().Interface("properties", object.Properties).Err(err).Msg("object has no pilot, using unitID as callsign")
		callsign = fmt.Sprintf("Unit %d", object.ID)
	}

	return &sim.Updated{
		Aircraft: trackfile.Aircraft{
			UnitID:    uint32(object.ID),
			Name:      callsign,
			Coalition: coalition,
			ACMIName:  name,
		},
		Frame: trackfile.Frame{
			Timestamp: time.Now(),
			Point:     coordinates.Location,
			Altitude:  coordinates.Altitude,
			Heading:   coordinates.Heading,
			Speed:     airspeed,
		},
	}, nil
}
