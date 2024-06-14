package webeditor

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/dharmab/skyeye/pkg/dcs"
	"github.com/dharmab/skyeye/pkg/encyclopedia"
	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/dharmab/skyeye/pkg/trackfile"
	"github.com/martinlindhe/unit"
)

const (
	BlueCoalitionName    = "blue"
	RedCoalitionName     = "red"
	NeutralCoalitionName = "neutrals"
)

func Load(mission Mission, updateCh chan<- dcs.Updated, bullseyeCh chan<- dcs.Bullseye) error {
	frameTime := time.Now()

	// TODO read terrain from mission JSON
	projector, error := dcs.NewProjector(encyclopedia.Caucases)
	if error != nil {
		return fmt.Errorf("error creating projector: %w", error)
	}

	coalitionMap := mission.Coalition
	for _, coalition := range []Coalition{coalitionMap.Blue, coalitionMap.Red} {
		var coalitionType types.Coalition
		if coalition.Name == BlueCoalitionName {
			coalitionType = types.CoalitionBlue
		} else if coalition.Name == RedCoalitionName {
			coalitionType = types.CoalitionRed
		}
		position, err := projector.Project(coalition.Bullseye.X, coalition.Bullseye.Y)
		if err != nil {
			return fmt.Errorf("error projecting bullseye for %s: %w", coalition.Name, err)
		}
		bullseye := dcs.Bullseye{
			Coalition: coalitionType,
			Point:     position,
		}
		bullseyeCh <- bullseye

		for _, country := range coalition.Country {
			for _, group := range country.Plane.Group {
				for _, plane := range group.Units {
					point, err := projector.Project(plane.X, plane.Y)
					if err != nil {
						slog.Error("Error projecting unit", "unit", plane, "error", err)
						continue
					}

					updated := dcs.Updated{
						Aircraft: trackfile.Aircraft{
							UnitID:     plane.UnitID,
							Name:       plane.Name,
							Coalition:  coalitionType,
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
