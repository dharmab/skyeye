package types

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/dharmab/skyeye/pkg/tacview/properties"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
)

type Object struct {
	ID         int
	properties map[string]string
	mut        sync.Mutex
}

func NewObject(id int) *Object {
	return &Object{
		ID:         id,
		properties: make(map[string]string),
	}
}

type Coordinates struct {
	Location orb.Point
	Altitude unit.Length
	// X is the object's native coordinate within the sim
	X float64
	// Y is the object's native coordiante within the sim
	Y     float64
	Roll  unit.Angle
	Pitch unit.Angle
	Yaw   unit.Angle
	// Heading is the object's flat earth heading
	Heading unit.Angle
}

func (o *Object) SetProperty(p, v string) {
	o.mut.Lock()
	defer o.mut.Unlock()
	o.properties[p] = v
}

func (o *Object) GetProperty(p string) (string, bool) {
	o.mut.Lock()
	defer o.mut.Unlock()
	val, ok := o.properties[p]
	if !ok {
		return "", false
	}
	return val, true
}

// GetTypes returns all object type tags
func (o *Object) GetTypes() ([]string, error) {
	val, ok := o.GetProperty(properties.Type)
	if !ok {
		return nil, errors.New("object does not contain types")
	}
	return strings.Split(val, "+"), nil
}

// GetCoordinates returns the coordinates of the object
// ref is the reference point from the global properties
func (o *Object) GetCoordinates(ref orb.Point) (Coordinates, error) {
	c := Coordinates{}

	val, ok := o.GetProperty(properties.Transform)
	if !ok {
		return c, errors.New("object does not contain coordinate transformation")
	}
	fields := strings.Split(val, "|")

	longitude, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return c, fmt.Errorf("error parsing longitude: %w", err)
	}
	latitude, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return c, fmt.Errorf("error parsing latitude: %w", err)
	}
	c.Location = orb.Point{longitude + ref.Lon(), latitude + ref.Lat()}

	if fields[2] != "" {
		altitude, err := strconv.ParseFloat(fields[2], 64)
		if err != nil {
			return c, fmt.Errorf("error parsing altitude: %w", err)
		}
		c.Altitude = unit.Length(altitude) * unit.Meter
	}

	var x, y, roll, pitch, yaw, heading string
	switch len(fields) {
	case 5:
		x = fields[3]
		y = fields[4]
	case 6:
		roll = fields[3]
		pitch = fields[4]
		yaw = fields[5]
	case 9:
		roll = fields[3]
		pitch = fields[4]
		yaw = fields[5]
		x = fields[6]
		y = fields[7]
		heading = fields[8]
	default:
		return c, fmt.Errorf("unexpected number of fields in coordinate transformation: %d", len(fields))
	}
	for s, fn := range map[string]func(float64){
		x:       func(v float64) { c.X = v },
		y:       func(v float64) { c.Y = v },
		roll:    func(v float64) { c.Roll = unit.Angle(v) * unit.Degree },
		pitch:   func(v float64) { c.Pitch = unit.Angle(v) * unit.Degree },
		yaw:     func(v float64) { c.Yaw = unit.Angle(v) * unit.Degree },
		heading: func(v float64) { c.Heading = unit.Angle(v) * unit.Degree },
	} {
		if s != "" {
			n, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return c, fmt.Errorf("error parsing native x: %w", err)
			}
			fn(n)
		}
	}
	return c, nil
}

func (o *Object) getNumericProperty(property string) (float64, error) {
	val, ok := o.GetProperty(property)
	if !ok {
		return 0, fmt.Errorf("object does not contain %s", property)
	}
	n, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing %s: %w", property, err)
	}
	return n, nil
}

func (o *Object) GetSpeed(property string) (unit.Speed, error) {
	n, err := o.getNumericProperty(property)
	if err != nil {
		return 0, err
	}
	return unit.Speed(n) * unit.MetersPerSecond, nil
}

func (o *Object) GetAngle(property string) (unit.Angle, error) {
	n, err := o.getNumericProperty(property)
	if err != nil {
		return 0, err
	}
	return unit.Angle(n) * unit.Degree, nil
}

func (o *Object) GetLength(property string) (unit.Length, error) {
	n, err := o.getNumericProperty(property)
	if err != nil {
		return 0, err
	}
	return unit.Length(n) * unit.Meter, nil
}
