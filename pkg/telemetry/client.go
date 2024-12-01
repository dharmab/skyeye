package telemetry

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

	"github.com/dharmab/goacmi/objects"
	"github.com/dharmab/goacmi/parsing"
	"github.com/dharmab/goacmi/properties"
	"github.com/dharmab/goacmi/tags"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/paulmach/orb"
	"github.com/rs/zerolog/log"
)

type client struct {
	starts chan sim.Started
	fades  chan sim.Faded

	// updateInterval is how often to send updates to the channels passed to Stream().
	updateInterval time.Duration
	// lastUpdateTime is the time that the last update was read.
	lastUpdateTime time.Time

	// referenceTime is the reference point provided in the ACMI data.
	referenceTime time.Time
	// referencePoint is the reference point provided in the ACMI data.
	referencePoint orb.Point
	// cursorTime is the current frame time, computed by adding the current time frame to the reference time.
	cursorTime time.Time

	// state maps object IDs to statuses.
	state map[uint64]*objects.Object
	// bullseyesIdx maps coalitions to bullseye object IDs.
	bullseyesIdx map[coalitions.Coalition]uint64
	// lock protects state and bullseyesIdx.
	lock sync.RWMutex
}

func NewClient(
	updateInterval time.Duration,
) *client {
	c := &client{
		starts:         make(chan sim.Started),
		fades:          make(chan sim.Faded),
		updateInterval: updateInterval,
	}
	c.reset()
	return c
}

func (c *client) Stream(ctx context.Context, wg *sync.WaitGroup, starts chan<- sim.Started, updates chan<- sim.Updated, fades chan<- sim.Faded) {
	ticker := time.NewTicker(c.updateInterval)
	defer ticker.Stop()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case start := <-c.starts:
				starts <- start
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case fade := <-c.fades:
				fades <- fade
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.sendUpdates(updates)
			}
		}
	}()

	<-ctx.Done()
}

func (c *client) Bullseye(coalition coalitions.Coalition) (orb.Point, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if id, ok := c.bullseyesIdx[coalition]; ok {
		if bullseye, ok := c.state[id]; ok {
			if coordinates, err := bullseye.GetCoordinates(c.referencePoint.Lon(), c.referencePoint.Lat()); err == nil {
				return orb.Point{*coordinates.Longitude, *coordinates.Latitude}, nil
			}
		}
	}
	return orb.Point{}, errors.New("bullseye not found")
}

func (c *client) Time() time.Time {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.cursorTime
}

func (c *client) sendUpdates(updates chan<- sim.Updated) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	for _, object := range c.state {
		logger := log.With().Uint64("id", object.ID).Logger()
		// Only send updates for aircraft.
		taglist, err := object.GetTypes()
		if err != nil {
			logger.Error().Err(err).Msg("error getting object types")
			continue
		}
		if !IsAircraft(taglist) {
			continue
		}
		logger = logger.With().Strs("tags", taglist).Logger()

		name, ok := object.GetProperty(properties.Name)
		if !ok {
			logger.Error().Msg("object missing name property")
			continue
		}
		logger = logger.With().Str("callsign", name).Logger()

		coordinates, err := object.GetCoordinates(c.referencePoint.Lon(), c.referencePoint.Lat())
		if err != nil {
			logger.Error().Err(err).Msg("error getting object coordinates")
			continue
		}

		callsign, ok := object.GetProperty(properties.Pilot)
		if !ok {
			// If the object has no pilot, use the object ID as the callsign.
			callsign = fmt.Sprintf("Unit %d", object.ID)
		}
		logger = logger.With().Str("callsign", callsign).Logger()

		acmiCoalition, ok := object.GetProperty(properties.Coalition)
		if !ok {
			logger.Error().Msg("object missing coalition property")
			continue
		}
		coalition := propertyToCoalition(acmiCoalition)

		frame := trackfiles.Frame{
			Time: c.cursorTime,
			Point: orb.Point{
				*coordinates.Longitude,
				*coordinates.Latitude,
			},
		}
		if coordinates.Altitude != nil {
			frame.Altitude = *coordinates.Altitude
		}
		if coordinates.Heading != nil {
			frame.Heading = *coordinates.Heading
		}
		if agl, err := object.GetLength(properties.AGL); err == nil {
			frame.AGL = &agl
		}

		updates <- sim.Updated{
			Labels: trackfiles.Labels{
				ID:        object.ID,
				Name:      callsign,
				Coalition: coalition,
				ACMIName:  name,
			},
			Frame: frame,
		}
	}
}

func (c *client) handleLines(ctx context.Context, reader *bufio.Reader) error {
	log.Info().Msg("resetting ACMI client state")
	c.reset()
	log.Info().Msg("sending mission start message")
	c.starts <- sim.Started{}

	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			gracePeriod := 10 * time.Minute
			if time.Since(c.lastUpdateTime) > gracePeriod {
				log.Warn().Time("lastUpdate", c.lastUpdateTime).Msg("stopped receiving updates")
				return errors.New("no updates received within grace period")
			}
		default:
			if err := c.handleUpdate(reader); err != nil {
				return fmt.Errorf("error reading ACMI stream: %w", err)
			}
		}
	}
}

func (c *client) handleUpdate(reader *bufio.Reader) error {
	line, err := reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("reached end of file: %w", err)
		}
		return fmt.Errorf("error reading line: %w", err)
	}

	if strings.HasSuffix(line, "\\\n") {
		line = line[:len(line)-2]
		for {
			next, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("error reading continuation line: %w", err)
			}
			line += next
			if !strings.HasSuffix(next, "\\\n") {
				break
			}
		}
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}
	if strings.HasPrefix(line, "//") {
		return nil
	}
	if line == fmt.Sprintf("%s=%s", properties.FileType, properties.ACMITacviewFileType) {
		return nil
	}
	if strings.HasPrefix(line, properties.FileVersion+"=") {
		return nil
	}

	if strings.HasPrefix(line, "#") {
		if err := c.handleTimeFrame(line); err != nil {
			return fmt.Errorf("error handling time frame: %w", err)
		}
		return nil
	}

	update, err := parsing.ParseObjectUpdate(line)
	if err != nil {
		return fmt.Errorf("error parsing object update: %w", err)
	}

	if update.ID == properties.GlobalObjectID {
		if err := c.updateGlobalObject(update); err != nil {
			return fmt.Errorf("error updating global object: %w", err)
		}
	} else {
		if err := c.updateObject(update); err != nil {
			return fmt.Errorf("error updating object: %w", err)
		}
	}

	c.lastUpdateTime = time.Now()
	return nil
}

func (c *client) handleTimeFrame(line string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if !strings.HasPrefix(line, "#") {
		return nil
	}
	offset, err := parsing.ParseTimeFrame(line)
	if err != nil {
		return fmt.Errorf("error parsing time frame: %w", err)
	}
	if c.referenceTime.IsZero() {
		return errors.New("time frame received before reference time")
	}
	c.cursorTime = c.referenceTime.Add(offset)
	return nil
}

func (c *client) updateGlobalObject(update *objects.Update) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if update.ID != properties.GlobalObjectID {
		return nil
	}

	if property, ok := update.Properties[properties.ReferenceTime]; ok {
		referenceTime, err := time.Parse(time.RFC3339, property)
		if err != nil {
			return fmt.Errorf("error parsing reference time: %w", err)
		}
		c.referenceTime = referenceTime
		if c.cursorTime.IsZero() {
			c.cursorTime = c.referenceTime
		}
	}

	if property, ok := update.Properties[properties.ReferenceLongitude]; ok {
		longitude, err := strconv.ParseFloat(property, 64)
		if err != nil {
			return fmt.Errorf("error parsing reference longitude: %w", err)
		}
		c.referencePoint = orb.Point{longitude, c.referencePoint.Lat()}
	}
	if property, ok := update.Properties[properties.ReferenceLatitude]; ok {
		latitude, err := strconv.ParseFloat(property, 64)
		if err != nil {
			return fmt.Errorf("error parsing reference latitude: %w", err)
		}
		c.referencePoint = orb.Point{c.referencePoint.Lon(), latitude}
	}

	return nil
}

func (c *client) updateObject(update *objects.Update) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	var isNewObject bool
	logger := log.With().Uint64("id", update.ID).Logger()

	object, ok := c.state[update.ID]
	if !ok {
		isNewObject = true
		object = objects.New(update.ID)
		c.state[update.ID] = object
	}

	if err := object.Update(update, c.referencePoint.Lon(), c.referencePoint.Lat()); err != nil {
		return fmt.Errorf("error updating object: %w", err)
	}

	taglist, err := object.GetTypes()
	if err != nil {
		return fmt.Errorf("attempted to update object of unknown type: %w", err)
	}

	logger = logger.With().Strs("tags", taglist).Logger()
	if callsign, ok := object.GetProperty(properties.Pilot); ok {
		logger = logger.With().Str("callsign", callsign).Logger()
	}
	if name, ok := object.GetProperty(properties.Name); ok {
		logger = logger.With().Str("aircraft", name).Logger()
	}
	if coalition, ok := object.GetProperty(properties.Coalition); ok {
		logger = logger.With().Stringer("coalition", propertyToCoalition(coalition)).Logger()
	}

	if isNewObject && IsRelevantObject(taglist) {
		logger.Info().Msg("recording new object")
	}

	isBullseye := slices.Contains(taglist, tags.Bullseye)
	if isBullseye {
		property, ok := object.GetProperty(properties.Coalition)
		if !ok {
			return errors.New("bullseye object missing coalition property")
		}
		coalition := propertyToCoalition(property)
		c.bullseyesIdx[coalition] = object.ID
	}

	if update.IsRemoval {
		delete(c.state, object.ID)
		c.fades <- sim.Faded{ID: object.ID}
		if IsRelevantObject(taglist) {
			logger.Info().Msg("recording object removal")
		}
	}

	return nil
}

func (c *client) reset() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.referenceTime = time.Time{}
	c.referencePoint = orb.Point{}
	c.cursorTime = time.Time{}
	c.state = map[uint64]*objects.Object{}
	c.bullseyesIdx = map[coalitions.Coalition]uint64{}
}

func IsAircraft(taglist []string) bool {
	return slices.Contains(taglist, tags.FixedWing) || slices.Contains(taglist, tags.Rotorcraft)
}

func IsBullseye(taglist []string) bool {
	return slices.Contains(taglist, tags.Bullseye)
}

func IsRelevantObject(taglist []string) bool {
	return IsAircraft(taglist) || IsBullseye(taglist)
}
