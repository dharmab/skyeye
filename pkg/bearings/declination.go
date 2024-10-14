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
var stubDate = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

// Declination returns the magnetic declination at the given point and time.
func Declination(p orb.Point, t time.Time) (unit.Angle, error) {
	if 1900 > t.Year() || t.Year() > 2025 {
		log.Warn().Time("date", t).Msg("year is outside IGRF model range, replacing with 2025")
		t = stubDate
	}
	field, err := igrfData.IGRF(p.Lat(), p.Lon(), 0, float64(t.Year())+float64(t.YearDay())/366)
	if err != nil {
		return 0, fmt.Errorf("failed to compute magnetic declination: %w", err)
	}
	return normalize(unit.Angle(field.Declination) * unit.Degree), nil
}
