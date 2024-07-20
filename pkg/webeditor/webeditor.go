package webeditor

import (
	"fmt"
	"time"

	"github.com/dharmab/skyeye/pkg/dcs"
	"github.com/dharmab/skyeye/pkg/encyclopedia"
	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/dharmab/skyeye/pkg/trackfile"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
)

func parseCoalition(coalitionName string) types.Coalition {
	if coalitionName == BlueCoalitionName {
		return types.CoalitionBlue
	}
	return types.CoalitionRed
}

func Load(mission Mission, updateCh chan<- dcs.Updated, bullseyeCh chan<- dcs.Bullseye) error {
	frameTime := time.Now()

	// TODO read terrain from mission JSON
	projector, error := dcs.NewProjector(encyclopedia.Caucases)
	if error != nil {
		return fmt.Errorf("error creating projector: %w", error)
	}

	coalitionMap := mission.Coalition
	for _, coalition := range []Coalition{coalitionMap.Blue, coalitionMap.Red} {
		position, err := projector.Project(coalition.Bullseye.X, coalition.Bullseye.Y)
		if err != nil {
			return fmt.Errorf("error projecting bullseye for %s: %w", coalition.Name, err)
		}
		bullseye := dcs.Bullseye{
			Coalition: parseCoalition(coalition.Name),
			Point:     position,
		}
		bullseyeCh <- bullseye

		for _, country := range coalition.Country {
			for _, group := range country.Plane.Group {
				for _, plane := range group.Units {
					point, err := projector.Project(plane.X, plane.Y)
					if err != nil {
						log.Error().Str("unit", plane.Name).Err(err).Msg("error projecting unit")
						continue
					}

					updated := dcs.Updated{
						Aircraft: trackfile.Aircraft{
							UnitID:     plane.UnitID,
							Name:       plane.Name,
							Coalition:  parseCoalition(coalition.Name),
							EditorType: plane.EditorType,
						},
						Frame: trackfile.Frame{
							Timestamp: frameTime,
							Point:     point,
							// Assuming altitude is ASL because I can't be arsed to implement AGL
							Altitude: unit.Length(plane.Altitude) * unit.Meter,
							Heading:  unit.Angle(plane.Heading) * unit.Degree,
							Speed:    unit.Speed(plane.Speed) * unit.KilometersPerHour,
						},
					}

					updateCh <- updated
				}
			}
		}
	}

	return nil
}
