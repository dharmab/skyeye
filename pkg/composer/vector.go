package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

func (c *Composer) ComposeVectorResponse(response brevity.VectorResponse) NaturalLanguageResponse {
	if !response.Contact {
		reply := response.Callsign + ", negative contact"
		return NaturalLanguageResponse{
			Subtitle: reply,
			Speech:   reply,
		}
	}

	if !response.Status {
		if response.Location == brevity.LocationTanker {
			reply := response.Callsign + ", no compatible tankers available"
			return NaturalLanguageResponse{
				Subtitle: reply,
				Speech:   reply,
			}
		}
		reply := response.Callsign + ", unable to provide vector to " + response.Location
		return NaturalLanguageResponse{
			Subtitle: reply,
			Speech:   reply,
		}
	}

	if response.BRA != nil {
		return c.composeTankerVectorResponse(response)
	}

	if !response.Vector.Bearing().IsMagnetic() {
		log.Error().Stringer("bearing", response.Vector.Bearing()).Msg("bearing provided to ComposeVectorResponse should be magnetic")
	}

	distance := int(response.Vector.Range().NauticalMiles())
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf(
			"%s, vector to %s, %s/%d",
			response.Callsign,
			response.Location,
			response.Vector.Bearing().String(),
			distance,
		),
		Speech: fmt.Sprintf(
			"%s, vector to %s, %s %d",
			response.Callsign,
			response.Location,
			pronounceBearing(response.Vector.Bearing()),
			distance,
		),
	}
}

func (c *Composer) composeTankerVectorResponse(response brevity.VectorResponse) NaturalLanguageResponse {
	if !response.BRA.Bearing().IsMagnetic() {
		log.Error().Stringer("bearing", response.BRA.Bearing()).Msg("bearing provided to composeTankerVectorResponse should be magnetic")
	}

	bearing := pronounceBearing(response.BRA.Bearing())
	_range := int(response.BRA.Range().NauticalMiles())
	altitude := c.composeAltitudeStacks(response.BRA.Stacks(), brevity.Unable)

	resp := NaturalLanguageResponse{
		Subtitle: fmt.Sprintf(
			"%s, nearest tanker, %s, BRA %s/%d, %s",
			response.Callsign,
			response.Location,
			response.BRA.Bearing().String(),
			_range,
			altitude,
		),
		Speech: fmt.Sprintf(
			"%s, nearest tanker, %s, bra %s, %d, %s",
			response.Callsign,
			response.Location,
			bearing,
			_range,
			altitude,
		),
	}

	if response.Track != brevity.UnknownDirection {
		track := fmt.Sprintf(", track %s", response.Track)
		resp.Subtitle += track
		resp.Speech += track
	}

	return resp
}
