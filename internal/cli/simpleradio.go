package cli

import (
	"github.com/dharmab/skyeye/pkg/simpleradio"
	"github.com/rs/zerolog/log"
)

func LoadFrequencies(frequencyStrs []string) []simpleradio.RadioFrequency {
	frequencies := make([]simpleradio.RadioFrequency, 0, len(frequencyStrs))
	for _, s := range frequencyStrs {
		freq, err := simpleradio.ParseRadioFrequency(s)
		if err != nil {
			log.Fatal().Err(err).Str("frequency", s).Msg("failed to parse SRS frequency")
		}
		frequencies = append(frequencies, *freq)
		log.Info().Stringer("frequency", freq).Msg("parsed SRS frequency")
	}
	return frequencies
}
