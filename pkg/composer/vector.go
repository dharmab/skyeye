package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

func (_ *Composer) ComposeVectorResponse(response brevity.VectorResponse) NaturalLanguageResponse {
	if !response.Contact {
		reply := response.Callsign + ", negative contact"
		return NaturalLanguageResponse{
			Subtitle: reply,
			Speech:   reply,
		}
	}

	if !response.Status {
		reply := response.Callsign + ", unable to provide vector to " + response.Location
		return NaturalLanguageResponse{
			Subtitle: reply,
			Speech:   reply,
		}
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
