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
	ID         uint64
	properties map[string]string
	mut        sync.RWMutex
}

func NewObject(id uint64) *Object {
	return &Object{
		ID:         id,
		properties: make(map[string]string),
	}
}

func (o *Object) Update(update *ObjectUpdate, ref orb.Point) error {
	var newTransform string
	for k, v := range update.Properties {
		if k == properties.Transform {
			newTransform = v
		} else {
			o.SetProperty(k, v)
		}
	}

	currentCoordinates, err := o.GetCoordinates(ref)
	if err != nil {
		o.SetProperty(properties.Transform, newTransform)
		//lint:ignore nilerr intentional overwrite of invalid coordinates
		return nil // ignore
	}
	err = currentCoordinates.Parse(newTransform, ref)
	if err != nil {
		return fmt.Errorf("error parsing coordinates: %w", err)
	}
	o.SetProperty(properties.Transform, currentCoordinates.Transform(ref))
	return nil
}

func (o *Object) SetProperty(p, v string) {
	o.mut.Lock()
	defer o.mut.Unlock()
	o.properties[p] = v
}

func (o *Object) GetProperty(p string) (string, bool) {
	val, ok := o.properties[p]
	if !ok {
		return "", false
	}
	return val, true
}

// GetTypes returns all object type tags.
func (o *Object) GetTypes() ([]string, error) {
	o.mut.RLock()
	defer o.mut.RUnlock()
	val, ok := o.GetProperty(properties.Type)
	if !ok {
		return nil, errors.New("object does not contain types")
	}
	return strings.Split(val, "+"), nil
}

// GetCoordinates returns the coordinates of the object, if possible.
// Many objects have insufficient information to determine their coordinates.
// In such a case, the function returns nil and no error.
// ref is the reference point from the global properties.
func (o *Object) GetCoordinates(ref orb.Point) (*Coordinates, error) {
	val, ok := o.GetProperty(properties.Transform)
	if !ok {
		return nil, errors.New("object does not contain coordinate transformation")
	}

	c := &Coordinates{}
	err := c.Parse(val, ref)
	if err != nil {
		return nil, fmt.Errorf("error parsing coordinates: %w", err)
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

func (o *Object) Properties() map[string]string {
	return o.properties
}
