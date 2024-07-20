package tacview

import (
	"bufio"
	"context"
	"errors"
	"fmt"
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
	"github.com/paulmach/orb"
	"github.com/rs/zerolog/log"
)

// https://www.tacview.net/documentation/acmi/en/

type ACMI interface {
	sim.Sim
	Start(context.Context) error
}

type streamer struct {
	acmi           *bufio.Scanner
	referencePoint orb.Point
	referenceTime  time.Time
	cursorTime     time.Time
	objects        map[int]types.Object
	fades          chan types.Object
	bullseye       orb.Point
}

func NewACMI(coalition coalitions.Coalition, acmi *bufio.Scanner) ACMI {
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
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if !s.acmi.Scan() {
				if s.acmi.Err() != nil {
					return fmt.Errorf("error scanning ACMI stream: %w", s.acmi.Err())
				}
			}
			line := s.acmi.Text()
			err := s.handleLine(line)
			if err != nil {
				log.Error().Err(err).Str("line", line).Msg("error handling ACMI line")
			}
		}
	}
}

func (s *streamer) handleLine(line string) error {
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
		object, ok := s.objects[update.ID]
		if ok {
			if _, ok := s.objects[update.ID]; ok {
				s.fades <- object
				delete(s.objects, update.ID)
				logger.Trace().Msg("object removed")
			}
			return nil
		}
	}

	t := update.Properties[properties.Type]
	if !strings.Contains(t, "FixedWing") && !strings.Contains(t, "Rotorcraft") {
		logger.Trace().Msg("object is not an aircraft")
		return nil
	}

	if _, ok := s.objects[update.ID]; !ok {
		logger.Debug().Msg("adding new object")
		s.objects[update.ID] = types.Object{ID: update.ID, Properties: make(map[string]string)}
	}
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
			log.Debug().Int("id", object.ID).Msg("object faded")
			fades <- sim.Faded{
				Timestamp: time.Now(),
				UnitID:    uint32(object.ID),
			}
		case <-ticker.C:
			for _, object := range s.objects {
				logger := log.With().Int("id", object.ID).Logger()
				types, err := object.GetTypes()
				if err != nil {
					logger.Error().Err(err).Msg("error getting object types")
					continue
				}
				if slices.Contains(types, tags.Bullseye) {
					err := s.updateBullseye(object)
					if err != nil {
						logger.Error().Err(err).Msg("error updating bullseye")
						continue
					}
					logger.Debug().Msg("bullseye updated")
				}
				if slices.Contains(types, tags.FixedWing) || slices.Contains(types, tags.Rotorcraft) {
					update, err := s.buildUpdate(object)
					if err != nil {
						logger.Error().Err(err).Msg("error building object update")
						continue
					}
					logger.Debug().Msg("aircraft updated")
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
	if !slices.Contains(types, tags.FixedWing) || !slices.Contains(types, tags.Bullseye) || !slices.Contains(types, tags.Rotorcraft) {
		return nil, err
	}
	callsign, err := object.GetProperty(properties.CallSign)
	if err != nil {
		return nil, err
	}
	name, err := object.GetProperty(properties.Name)
	if err != nil {
		return nil, err
	}
	coordinates, err := object.GetCoordinates(s.referencePoint)
	if err != nil {
		return nil, err
	}
	heading, err := object.GetAngle(properties.HDG)
	if err != nil {
		return nil, err
	}
	airspeed, err := object.GetSpeed(properties.TAS)
	if err != nil {
		return nil, err
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

	return &sim.Updated{
		Aircraft: trackfile.Aircraft{
			UnitID:     uint32(object.ID),
			Name:       callsign,
			Coalition:  coalition,
			EditorType: "",
			ACMIName:   name,
		},
		Frame: trackfile.Frame{
			Timestamp: time.Now(),
			Point:     coordinates.Location,
			Altitude:  coordinates.Altitude,
			Heading:   heading,
			Speed:     airspeed,
		},
	}, nil
}
