package types

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/rs/zerolog/log"
)

type Coordinates struct {
	// Flag indicating the presence or absence of the related field
	ValidLon, ValidLat bool
	// Location is the object's lon/lat position
	Location orb.Point
	// Altitude above sea level
	Altitude *unit.Length
	// X is the object's native coordinate within the sim
	X *float64
	// Y is the object's native coordiante within the sim
	Y *float64
	// Roll angle
	Roll *unit.Angle
	// Pitch angle
	Pitch *unit.Angle
	// Yaw angle (this may be different from heading)
	Yaw *unit.Angle
	// Heading is the object's flat earth heading
	Heading *unit.Angle
}

func NewCoordinates(
	point orb.Point,
	validLon, validLat bool,
	altitude *unit.Length,
	x, y *float64,
	roll, pitch, yaw, heading *unit.Angle,
) *Coordinates {
	return &Coordinates{
		Location: point,
		ValidLon: validLon,
		ValidLat: validLat,
		Altitude: altitude,
		X:        x,
		Y:        y,
		Roll:     roll,
		Pitch:    pitch,
		Yaw:      yaw,
		Heading:  heading,
	}
}

func (c *Coordinates) Update(next *Coordinates) {
	longitude := c.Location.Lon()
	if next.ValidLon {
		longitude = next.Location.Lon()
		c.ValidLon = true
	}
	latitude := c.Location.Lat()
	if next.ValidLat {
		latitude = next.Location.Lat()
		c.ValidLat = true
	}
	c.Location = orb.Point{longitude, latitude}

	if next.Altitude != nil {
		c.Altitude = next.Altitude
	}
	if next.X != nil {
		c.X = next.X
	}
	if next.Y != nil {
		c.Y = next.Y
	}
	if next.Roll != nil {
		c.Roll = next.Roll
	}
	if next.Pitch != nil {
		c.Pitch = next.Pitch
	}
	if next.Yaw != nil {
		c.Yaw = next.Yaw
	}
	if next.Heading != nil {
		c.Heading = next.Heading
	}
}

func (c *Coordinates) Parse(transform string, ref orb.Point) error {
	fields := strings.Split(transform, "|")

	logger := log.With().Str("transform", transform).Logger()
	if len(fields) < 3 {
		logger.Trace().Msg("unexpected number of fields in coordinate transformation")
		return nil
	}

	var longitude, latitude float64
	if fields[0] != "" {
		offset, err := strconv.ParseFloat(fields[0], 64)
		if err != nil {
			logger.Trace().Err(err).Msg("error parsing longitude")
		}
		longitude = ref.Lon() + offset
		c.ValidLon = true
	}
	if fields[1] != "" {
		offset, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			logger.Trace().Err(err).Msg("error parsing latitude")
		}
		latitude = ref.Lat() + offset
		c.ValidLat = true
	}
	c.Location = orb.Point{longitude, latitude}

	var alt, x, y, roll, pitch, yaw, heading string
	switch len(fields) {
	case 3:
		alt = fields[2]
	case 5:
		alt = fields[2]
		x = fields[3]
		y = fields[4]
	case 6:
		alt = fields[2]
		roll = fields[3]
		pitch = fields[4]
		yaw = fields[5]
	case 9:
		alt = fields[2]
		roll = fields[3]
		pitch = fields[4]
		yaw = fields[5]
		x = fields[6]
		y = fields[7]
		heading = fields[8]
	default:
		log.Error().Str("transform", transform).Msg("unexpected number of fields in coordinate transformation")
		return fmt.Errorf("unexpected number of fields in coordinate transformation: %d", len(fields))
	}

	for s, fn := range map[string]func(float64){
		alt: func(v float64) {
			a := unit.Length(v) * unit.Meter
			c.Altitude = &a
		},
		x: func(v float64) {
			c.X = &v
		},
		y: func(v float64) {
			c.Y = &v
		},
		roll: func(v float64) {
			θ := unit.Angle(v) * unit.Degree
			c.Roll = &θ
		},
		pitch: func(v float64) {
			θ := unit.Angle(v) * unit.Degree
			c.Pitch = &θ
		},
		yaw: func(v float64) {
			θ := unit.Angle(v) * unit.Degree
			c.Yaw = &θ
		},
		heading: func(v float64) {
			θ := unit.Angle(v) * unit.Degree
			c.Heading = &θ
		},
	} {
		if s != "" {
			n, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return fmt.Errorf("error parsing native transform value: %w", err)
			}
			fn(n)
		}
	}

	return nil
}

func (c *Coordinates) Transform(ref orb.Point) string {
	fields := make([]string, 9)
	if c.ValidLon {
		fields[0] = fmt.Sprintf("%f", c.Location.Lon()-ref.Lon())
	}
	if c.ValidLat {
		fields[1] = fmt.Sprintf("%f", c.Location.Lat()-ref.Lat())
	}
	if c.Altitude != nil {
		fields[2] = fmt.Sprintf("%f", c.Altitude.Meters())
	}
	if c.Roll != nil {
		fields[3] = fmt.Sprintf("%f", c.Roll.Degrees())
	}
	if c.Pitch != nil {
		fields[4] = fmt.Sprintf("%f", c.Pitch.Degrees())
	}
	if c.Yaw != nil {
		fields[5] = fmt.Sprintf("%f", c.Yaw.Degrees())
	}
	if c.X != nil {
		fields[6] = fmt.Sprintf("%f", *c.X)
	}
	if c.Y != nil {
		fields[7] = fmt.Sprintf("%f", *c.Y)
	}
	if c.Heading != nil {
		fields[8] = fmt.Sprintf("%f", c.Heading.Degrees())
	}
	return strings.Join(fields, "|")
}
