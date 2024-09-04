package simpleradio

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
)

// RadioFrequency selects a frequency and either AM or FM modulation.
type RadioFrequency struct {
	Frequency  unit.Frequency
	Modulation types.Modulation
}

// ParseRadioFrequency parses a string into a RadioFrequency.
// The string should be a postive decimal number optionally followed by either "AM" or "FM".
// If the modulation is not recognized, it defaults to AM.
func ParseRadioFrequency(s string) (*RadioFrequency, error) {
	pos := strings.IndexFunc(s, func(r rune) bool {
		return (r < '0' || r > '9') && r != '.' && r != '-'
	})

	var prefix, suffix string
	if pos == -1 {
		prefix = s
	} else {
		prefix = s[:pos]
		suffix = strings.TrimSpace(s[pos:])
	}

	mhz, err := strconv.ParseFloat(prefix, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frequency: %w", err)
	}
	if math.IsNaN(mhz) || math.IsInf(mhz, 0) || mhz <= 0 {
		return nil, errors.New("frequency must be a real postive number")
	}
	frequency := unit.Frequency(mhz) * unit.Megahertz

	var modulation types.Modulation
	switch suffix {
	case "FM":
		modulation = types.ModulationFM
	case "AM":
		modulation = types.ModulationAM
	default:
		log.Warn().Str("input", s).Msg("unknown modulation, defaulting to AM")
		modulation = types.ModulationAM
	}

	return &RadioFrequency{
		Frequency:  frequency,
		Modulation: modulation,
	}, nil
}

func (f RadioFrequency) IsSameFrequency(other RadioFrequency) bool {
	return f.Frequency == other.Frequency && f.Modulation == other.Modulation
}

// String representation of the RadioFrequency.
func (f RadioFrequency) String() string {
	var suffix string
	switch f.Modulation {
	case types.ModulationFM:
		suffix = "FM"
	case types.ModulationAM:
		suffix = "AM"
	}

	return fmt.Sprintf("%f.3%s", f.Frequency, suffix)
}
