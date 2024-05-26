package dcs

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/encyclopedia"
	"github.com/paulmach/orb"
	"github.com/xeonx/geom"
	proj "github.com/xeonx/proj4"
)

type Projector interface {
	Project(x float64, y float64) (orb.Point, error)
}

type projector struct {
	transformation proj.Transformation
}

func NewProjector(terrain encyclopedia.Terrain) (*projector, error) {
	// Convert from DCS transverse Mercator X/Y in meters
	defintion := fmt.Sprintf(
		// Reference https://proj.org/en/9.4/operations/projections/tmerc.html#parameters for explanation of parameters
		// Note we're using a very old version of PROJ (v4) so we have to provide +no_defs (https://proj.org/en/9.4/usage/differences.html#removal-of-proj-def-dat)
		"+proj=tmerc +lat_0=0 +lon_0=%f +k_0=0.9996 +x_0=%f +y_0=%f +towgs84=0,0,0,0,0,0,0 +units=m +vunits=m +ellps=WGS84 +no_defs +axis=neu",
		terrain.CentralMeridian,
		terrain.FalseEasting,
		terrain.FalseNorthing,
	)
	sourceProjection, err := proj.InitPlus(defintion)
	if err != nil {
		return nil, fmt.Errorf("error initializing simulator coordinate projection: %w", err)
	}

	// Convert to simple Long/Lat projection
	destProjection, err := proj.InitPlus("+proj=longlat +datum=WGS84 +no_defs")
	if err != nil {
		return nil, fmt.Errorf("error initializing long/lat projection: %w", err)
	}

	transformation, err := proj.NewTransformation(sourceProjection, destProjection)
	if err != nil {
		return nil, fmt.Errorf("error creating transformation: %w", err)
	}

	return &projector{transformation: transformation}, nil
}

func (p *projector) Project(x float64, y float64) (orb.Point, error) {
	point := geom.Point{X: x, Y: y}
	points := []geom.Point{point}
	err := p.transformation.TransformPoints(points)
	if err != nil {
		return orb.Point{}, fmt.Errorf("error transforming point (%f, %f): %w", x, y, err)
	}
	return orb.Point{
		points[0].X,
		points[0].Y,
	}, nil
}
