package webeditor

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/dharmab/skyeye/pkg/dcs"
	"github.com/dharmab/skyeye/pkg/encyclopedia"
	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/dharmab/skyeye/pkg/trackfile"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
)

const (
	BlueCoalitionName    = "blue"
	RedCoalitionName     = "red"
	NeutralCoalitionName = "neutral"
)

func LoadJSON(mission map[string]json.RawMessage, updateCh chan<- dcs.Updated, bullseyeCh chan<- orb.Point) error {
	frameTime := time.Now()

	// TODO read terrain from mission JSON
	projector, error := dcs.NewProjector(encyclopedia.Caucases)
	if error != nil {
		return fmt.Errorf("error creating projector: %w", error)
	}

	var coalitions map[string]json.RawMessage
	if err := json.Unmarshal(mission["coalition"], &coalitions); err != nil {
		return fmt.Errorf("error unmarshalling top-level key 'coalition': %w", err)
	}
	for coalitionName, coalitionJson := range coalitions {
		var coalitionID types.Coalition
		if coalitionName == BlueCoalitionName {
			coalitionID = types.CoalitionBlue
		} else if coalitionName == RedCoalitionName {
			coalitionID = types.CoalitionRed
		}

		var coalition map[string]json.RawMessage
		if err := json.Unmarshal(coalitionJson, &coalition); err != nil {
			return fmt.Errorf("error unmarshalling coalition[%q]: %w", coalitionName, err)
		}

		var bullseye map[string]map[string]float64
		if err := json.Unmarshal(coalition["bullseye"], &bullseye); err != nil {
			return fmt.Errorf("error unmarshalling coalition[%s].bullseye: %w", coalitionName, err)
		}

		var countries []json.RawMessage
		if err := json.Unmarshal(coalition["country"], &countries); err != nil {
			return fmt.Errorf("error unmarshalling coalition[%s].country[]': %w", coalitionName, err)
		}
		for i, countryElement := range countries {
			var country map[string]json.RawMessage
			if err := json.Unmarshal(countryElement, &country); err != nil {
				return fmt.Errorf("error unmarshalling coalition[%s].country[%d]': %w", coalitionName, i, err)
			}
			var planeClass map[string]json.RawMessage
			if err := json.Unmarshal(country["plane"], &planeClass); err != nil {
				return fmt.Errorf("error unmarshalling coalition[%s].country[%d].plane': %w", coalitionName, i, err)
			}
			var planeGroups []json.RawMessage
			if err := json.Unmarshal(planeClass["group"], &planeGroups); err != nil {
				return fmt.Errorf("error unmarshalling coalition[%s].country[%d].plane.group[]': %w", coalitionName, i, err)
			}
			for j, groupJson := range planeGroups {
				var group map[string]json.RawMessage
				if err := json.Unmarshal(groupJson, &group); err != nil {
					return fmt.Errorf("error unmarshalling coalition[%s].country[%d].plane.group[%d]': %w", coalitionName, i, j, err)
				}
				var units []json.RawMessage
				if err := json.Unmarshal(group["units"], &units); err != nil {
					return fmt.Errorf("error unmarshalling coalition[%q].country[%d].plane.group[%d].units[]': %w", coalitionName, i, j, err)
				}

				for k, unitJson := range units {
					var simUnit map[string]json.RawMessage
					if err := json.Unmarshal(unitJson, &simUnit); err != nil {
						return fmt.Errorf("error unmarshalling coalition[%q].country[%d].plane.group[%d].units[%d]': %w", coalitionName, i, j, k, err)
					}

					var unitID uint32
					if err := json.Unmarshal(simUnit["unitID"], &unitID); err != nil {
						return fmt.Errorf("error unmarshalling coalition[%q].country[%d].plane.group[%d].units[%d].unitID': %w", coalitionName, i, j, k, err)
					}
					var name string
					if err := json.Unmarshal(simUnit["name"], &name); err != nil {
						return fmt.Errorf("error unmarshalling coalition[%q].country[%d].plane.group[%d].units[%d].name': %w", coalitionName, i, j, k, err)
					}
					var editorType string
					if err := json.Unmarshal(simUnit["type"], &editorType); err != nil {
						return fmt.Errorf("error unmarshalling coalition[%q].country[%d].plane.group[%d].units[%d].type': %w", coalitionName, i, j, k, err)
					}
					var x float64
					if err := json.Unmarshal(simUnit["x"], &x); err != nil {
						return fmt.Errorf("error unmarshalling coalition[%q].country[%d].plane.group[%d].units[%d].x': %w", coalitionName, i, j, k, err)
					}
					var y float64
					if err := json.Unmarshal(simUnit["y"], &y); err != nil {
						return fmt.Errorf("error unmarshalling coalition[%q].country[%d].plane.group[%d].units[%d].y': %w", coalitionName, i, j, k, err)
					}
					var altitude float64
					if err := json.Unmarshal(simUnit["alt"], &altitude); err != nil {
						return fmt.Errorf("error unmarshalling coalition[%q].country[%d].plane.group[%d].units[%d].alt': %w", coalitionName, i, j, k, err)
					}
					var heading float64
					if err := json.Unmarshal(simUnit["heading"], &heading); err != nil {
						return fmt.Errorf("error unmarshalling coalition[%q].country[%d].plane.group[%d].units[%d].heading': %w", coalitionName, i, j, k, err)
					}
					var speed float64
					if err := json.Unmarshal(simUnit["speed"], &speed); err != nil {
						return fmt.Errorf("error unmarshalling coalition[%q].country[%d].plane.group[%d].units[%d].speed': %w", coalitionName, i, j, k, err)
					}

					point, err := projector.Project(x, y)
					if err != nil {
						slog.Error("Error projecting unit", "unit", simUnit, "error", err)
						continue
					}

					updated := dcs.Updated{
						Aircraft: trackfile.Aircraft{
							UnitID:     unitID,
							Name:       name,
							Coalition:  coalitionID,
							EditorType: editorType,
						},
						Frame: trackfile.Frame{
							Timestamp: frameTime,
							// Convert from DCS in-game coordinates to Long/Lat
							// https://github.com/DCS-Web-Editor/dcs-web-editor-mono/blob/main/packages/map-projection/src/index.ts
							Point: point,
							// Assuming altitude is ASL because I can't be arsed to implement AGL
							Altitude: unit.Length(altitude) * unit.Meter,
							Heading:  unit.Angle(heading) * unit.Degree,
							Speed:    unit.Speed(speed) * unit.MetersPerSecond,
						},
					}
					updateCh <- updated
				}
			}
		}
	}
	return nil
}
