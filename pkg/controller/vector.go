package controller

import (
	"context"
	"slices"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/locations"
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/rs/zerolog/log"
)

// HandleVector handles a VECTOR request by computing the bearing and distance from the requesting aircraft to a named location.
func (c *Controller) HandleVector(ctx context.Context, request *brevity.VectorRequest) {
	logger := log.With().Str("callsign", request.Callsign).Type("type", request).Logger()
	logger.Debug().Msg("handling request")

	response := brevity.VectorResponse{
		Callsign: request.Callsign,
		Location: request.Location,
	}

	var trackfile *trackfiles.Trackfile
	response.Callsign, trackfile, response.Contact = c.findCallsign(request.Callsign)

	var targetLocation *locations.Location
	for _, location := range c.locations {
		if location.Names == nil {
			continue
		}
		if slices.Contains(location.Names, request.Location) {
			targetLocation = &location
			break
		}
	}
	response.Status = targetLocation != nil

	if response.Contact && response.Status {
		origin := trackfile.LastKnown().Point
		target := targetLocation.Point()
		declination := c.scope.Declination(origin)
		bearing := spatial.TrueBearing(origin, target).Magnetic(declination)
		distance := spatial.Distance(origin, target)
		response.Vector = brevity.NewVector(bearing, distance)
	}

	c.calls <- NewCall(ctx, response)
}
