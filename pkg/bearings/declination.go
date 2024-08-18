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
	if t.Year() < 1900 {
		log.Warn().Msg("date is too early for IGRF model, replacing with real-time date")
		t = time.Now()
	}
	if t.Year() > 2025 {
		log.Warn().Msg("year is too late for IGRF model, replacing with 2025")
		t = time.Date(2025, t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	}
	field, err := igrfData.IGRF(p.Lat(), p.Lon(), 0, float64(t.Year())+float64(t.YearDay())/366)
	if err != nil {
		return 0, fmt.Errorf("failed to compute magnetic declination: %w", err)
	}
	return normalize(unit.Angle(field.Declination) * unit.Degree), nil
}
