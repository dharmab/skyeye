package bearings

import (
	"fmt"
	"time"

	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/proway2/go-igrf/igrf"
	"github.com/rs/zerolog/log"
)

var igrfData = igrf.New()

// Declination returns the magnetic declination at the given point and time.
func Declination(p orb.Point, t time.Time) (unit.Angle, error) {
	if t.Year() > 2025 {
		log.Warn().Msg("clamping date to 2025 for purposes of computing magnetic declination due to IGRF model limits")
		t = time.Date(2025, t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	}
	field, err := igrfData.IGRF(p.Lat(), p.Lon(), 0, float64(t.Year())+float64(t.YearDay())/366)
	if err != nil {
		return 0, fmt.Errorf("failed to compute magnetic declination: %w", err)
	}
	return Normalize(unit.Angle(field.Declination) * unit.Degree), nil
}
